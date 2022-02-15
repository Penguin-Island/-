package be

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memcached"
	"github.com/gin-gonic/gin"
	"github.com/memcachier/mc"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func isFlagEnabled(flags []string, key string) bool {
	for _, k := range flags {
		if k == key {
			return true
		}
	}
	return false
}

func launchWebpackServer() error {
	cmd := exec.Command("npm", "install")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("npm", "run", "_server")
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

func initDatabase() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password= dbname=ohatori port=5432 sslmode=disable TimeZone=Asia/Tokyo"
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func Run() {
	if isFlagEnabled(os.Args[1:], "release") {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	log.SetReportCaller(true)

	db, err := initDatabase()
	if err != nil {
		log.Fatal(err)
	}
	_ = db

	r := gin.Default()

	client := mc.NewMC("localhost:11211", "", "")
	store := memcached.NewMemcacheStore(client, "", []byte(""))
	r.Use(sessions.Sessions("session", store))

	if isFlagEnabled(os.Args[1:], "noproxy") {
		r.SetTrustedProxies([]string{})
		if err := launchWebpackServer(); err != nil {
			log.Fatal(err)
		}
		r.NoRoute(forwardToWebpack)
	} else {
		r.SetTrustedProxies([]string{"127.0.0.1"})
	}

	r.GET("/hoge", func(c *gin.Context) {
		sess := sessions.Default(c)
		var count int
		v := sess.Get("count")
		if v == nil {
			count = 0
		} else {
			count = v.(int)
			fmt.Println(count)
			count++
		}
		sess.Set("count", count)
		sess.Save()
	})

	r.Run("0.0.0.0:8000")
}
