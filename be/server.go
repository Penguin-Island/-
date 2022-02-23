package be

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/gin-contrib/sessions"
	redisSess "github.com/gin-contrib/sessions/redis"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Member struct {
	gorm.Model
	UserName string `gorm:"unique"`
	Password string
	GroupId  uint
}

type App struct {
	db         *gorm.DB
	gameStates GameStates
	redis      *redis.Client
}

func getRedisURL() (*url.URL, error) {
	url, err := url.Parse(os.Getenv("REDIS_URL"))
	if err != nil {
		return nil, err
	}
	return url, nil
}

func NewApp() *App {
	app := new(App)
	app.gameStates.communicators = make(map[uint]chan InternalNotification)
	return app
}

func isFlagEnabled(flags []string, key string) bool {
	for _, k := range flags {
		if k == key {
			return true
		}
	}
	return false
}

func launchWebpackServer(runNpmInstall bool) error {
	if runNpmInstall {
		cmd := exec.Command("npm", "install")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	cmd := exec.Command("npm", "run", "_server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	go func() {
		if err := cmd.Run(); err != nil {
			log.Println(err)
		}
		fmt.Println()
	}()
	return nil
}

func forwardToWebpack(c *gin.Context) {
	c.Request.URL.Host = "localhost:8080"
	c.Request.URL.Scheme = "http"
	c.Request.RequestURI = ""
	resp, err := http.DefaultClient.Do(c.Request)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(500)
		return
	}
	for k, vs := range resp.Header {
		for _, v := range vs {
			c.Header(k, v)
		}
	}
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

func initDatabase(verboseLog bool) (*gorm.DB, error) {
	config := gorm.Config{}
	if verboseLog {
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &config)
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Member{}); err != nil {
		log.Warn(err)
	}
	if err := db.AutoMigrate(&Group{}); err != nil {
		log.Warn(err)
	}
	if err := db.AutoMigrate(&Invitation{}); err != nil {
		log.Warn(err)
	}
	if err := db.AutoMigrate(&Statistics{}); err != nil {
		log.Warn(err)
	}
	return db, nil
}

func Run() {
	err := godotenv.Load()
	if err != nil {
		log.Info("Error loading .env file (actual environment variables will be used)")
	}

	isDebugMode := isFlagEnabled(os.Args[1:], "debug")

	if isDebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	log.SetReportCaller(true)

	app := NewApp()

	db, err := initDatabase(isDebugMode)
	if err != nil {
		log.Fatal(err)
	}
	app.db = db

	redisUrl, err := url.Parse(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal(err)
	}
	redisPassword, _ := redisUrl.User.Password()

	app.redis = redis.NewClient(&redis.Options{
		Addr:     redisUrl.Host,
		Password: redisPassword,
	})

	rand.Seed(time.Now().Unix())

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		if !isDebugMode && c.GetHeader("X-Forwarded-Proto") == "http" {
			allowedUrl, err := url.Parse(os.Getenv("ALLOWED_ORIGIN"))
			if err != nil {
				log.Error(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			url := c.Request.URL
			url.Scheme = "https"
			url.Host = allowedUrl.Host
			c.Redirect(http.StatusMovedPermanently, url.String())
			c.Abort()
		}
	})

	r.Use(func(c *gin.Context) {
		if c.Request.Method == http.MethodPost {
			if c.GetHeader("Origin") == os.Getenv("ALLOWED_ORIGIN") {
				return
			}
		} else if c.Request.Method == http.MethodGet {
			if c.GetHeader("Upgrade") == "websocket" {
				if c.GetHeader("Origin") == os.Getenv("ALLOWED_ORIGIN") {
					return
				}
			} else {
				return
			}
		} else {
			return
		}
		c.Status(http.StatusBadGateway)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Writer.WriteString(`<!doctype html>
<html>
<head>
    <title>502 Bad Gateway</title>
</head>
<body>
    <h1>502 Bad Gateway</h1>
</body>
</html>
`)
		c.Abort()
	})

	store, err := redisSess.NewStore(10, "tcp", redisUrl.Host, redisPassword, []byte(os.Getenv("SESSION_SECRET")))
	if err != nil {
		log.Fatal(err)
	}
	redisSess.SetKeyPrefix(store, "session:")
	r.Use(sessions.Sessions("session", store))

	if isFlagEnabled(os.Args[1:], "noproxy") {
		r.SetTrustedProxies([]string{})
	} else {
		r.SetTrustedProxies([]string{"127.0.0.1"})
	}

	var staticHandler func(*gin.Context)
	if isDebugMode {
		staticHandler = forwardToWebpack
		if err := launchWebpackServer(!isFlagEnabled(os.Args[1:], "nonpminstall")); err != nil {
			log.Fatal(err)
		}
		r.NoRoute(forwardToWebpack)
	} else {
		staticHandler = static.Serve("/", static.LocalFile("dist", false))
		r.NoRoute(staticHandler)
	}

	r.GET("/", func(c *gin.Context) {
		sess := sessions.Default(c)
		userId := sess.Get("user_id")
		if userId != nil {
			c.Redirect(http.StatusFound, "/game/")
			return
		}
		staticHandler(c)
	})

	r.POST("/users/login", func(c *gin.Context) {
		handleLogin(app, c)
	})

	r.GET("/finish/", func(c *gin.Context) {
		sess := sessions.Default(c)
		userId := sess.Get("user_id")
		if userId == nil {
			c.Redirect(http.StatusFound, "/")
			return
		}
		staticHandler(c)
	})

	r.GET("/game/", func(c *gin.Context) {
		sess := sessions.Default(c)
		userId := sess.Get("user_id")
		if userId == nil {
			c.Redirect(http.StatusFound, "/")
			return
		}
		staticHandler(c)
	})

	r.GET("/game_ws", func(c *gin.Context) {
		handleSocketConnection(app, c)
	})

	r.POST("/users/new", func(c *gin.Context) {
		handleRegisterUser(app, c)
	})

	r.GET("/users/find", func(c *gin.Context) {
		handleFindUser(app, c)
	})

	r.GET("/logout", func(c *gin.Context) {
		sess := sessions.Default(c)
		sess.Clear()
		sess.Save()
		c.Redirect(http.StatusFound, "/")
	})

	r.GET("/users/info", func(c *gin.Context) {
		handleGetUserInfo(app, c)
	})

	r.GET("/users/statistics", func(c *gin.Context) {
		handleGetStatistics(app, c)
	})

	r.POST("/groups/invite", func(c *gin.Context) {
		handleInvite(app, c)
	})

	r.GET("/groups/invitations", func(c *gin.Context) {
		handleGetInvitations(app, c)
	})

	r.POST("/groups/decline_invitation", func(c *gin.Context) {
		handleDeclineInvitations(app, c)
	})

	r.POST("/groups/join", func(c *gin.Context) {
		handleJoin(app, c)
	})

	r.POST("/groups/unjoin", func(c *gin.Context) {
		handleUnjoin(app, c)
	})

	r.POST("/groups/wake_up_time", func(c *gin.Context) {
		handleSetTime(app, c)
	})

	log.Fatal(r.Run(fmt.Sprintf(":%s", os.Getenv("PORT"))))
}
