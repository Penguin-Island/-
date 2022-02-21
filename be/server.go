package be

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memcached"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/memcachier/mc"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"golang.org/x/crypto/bcrypt"
)

type Member struct {
	gorm.Model
	PlayerTag string `gorm:"unique"`
	UserName  string
	Password  string
	GroupId   uint
}

type App struct {
	db         *gorm.DB
	gameStates GameStates
	memcached  *mc.Client
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

func generatePlayerTag(userName string) string {
	return userName + strconv.Itoa(rand.Intn(8999)+1000)
}

func initDatabase(verboseLog bool) (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password= dbname=ohatori port=5432 sslmode=disable TimeZone=Asia/Tokyo"

	config := gorm.Config{}
	if verboseLog {
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), &config)
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
	if isFlagEnabled(os.Args[1:], "release") {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	log.SetReportCaller(true)

	app := NewApp()

	db, err := initDatabase(!isFlagEnabled(os.Args[1:], "release"))
	if err != nil {
		log.Fatal(err)
	}
	app.db = db

	rand.Seed(time.Now().Unix())

	r := gin.Default()

	app.memcached = mc.NewMC("localhost:11211", "", "")
	store := memcached.NewMemcacheStore(app.memcached, "session-", []byte(""))
	r.Use(sessions.Sessions("session", store))

	if isFlagEnabled(os.Args[1:], "noproxy") {
		r.SetTrustedProxies([]string{})
	} else {
		r.SetTrustedProxies([]string{"127.0.0.1"})
	}

	var staticHandler func(*gin.Context)
	if isFlagEnabled(os.Args[1:], "release") {
		staticHandler = static.Serve("/", static.LocalFile("dist", false))
		r.NoRoute(staticHandler)
	} else {
		staticHandler = forwardToWebpack
		if err := launchWebpackServer(!isFlagEnabled(os.Args[1:], "nonpminstall")); err != nil {
			log.Fatal(err)
		}
		r.NoRoute(forwardToWebpack)
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

	r.POST("/", func(c *gin.Context) {
		var member Member
		playerTag := c.PostForm("playerTag")
		password := c.PostForm("password")
		if err := db.First(&member, "player_tag = ?", playerTag).Error; err != nil {
			log.Error(err)
			c.Redirect(http.StatusFound, "/")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(password)); err != nil {
			log.Error(err)
			c.Redirect(http.StatusFound, "/")
			return
		}
		sess := sessions.Default(c)
		sess.Set("user_id", member.ID)
		sess.Save()
		c.Redirect(http.StatusFound, "/game/")
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

	if err := r.Run("0.0.0.0:8000"); err != nil {
		if !isFlagEnabled(os.Args[1:], "release") {
			log.Println(err)
			log.Println("fallback to :1333")

			r.Run("0.0.0.0:1333")
		} else {
			log.Fatal(err)
		}
	}
}
