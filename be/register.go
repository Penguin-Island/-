package be

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func isValidUserName(userName string) bool {
	for _, r := range []rune(userName) {
		if !('a' <= r && r <= 'z') && !('A' <= r && r <= 'Z') && !('0' <= r && r <= '9') && ( r != '_' ) && ( r != '-') {
			return false
		}
	}
	return true
}

func isValidPassword(password string) bool {
	hasAlpha := false
	hasDigit := false
	for _, r := range []rune(password) {
		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') {
			hasAlpha = true
		} else if '0' <= r && r <= '9' {
			hasDigit = true
		}
	}
	return hasAlpha && hasDigit
}

func registerUser(app *App, userName, password string) (bool, error) {
	if !(isValidUserName(userName) && isValidPassword(password)) {
		return false, nil
	}


	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}

	member := Member{
		PlayerTag: generatePlayerTag(userName),
		UserName:  userName,
		Password:  string(hashed),
	}
	for i := 0; i < 100; i++ {
		if err = app.db.Create(&member).Error; err != nil {
			log.Println(err)
			member.PlayerTag = generatePlayerTag(userName)
			continue
		}
		break
	}

	if err != nil {
		return true, err
	}
	return true, nil
}

func handleRegisterUser(app *App, c *gin.Context) {
	userName := c.PostForm("username")
	password := c.PostForm("password")

	if len(userName) < 3 || len(password) < 10 {
		c.Redirect(http.StatusFound, "/register/")
		return
	}

	if acceptable, err := registerUser(app, userName, password); !acceptable {
		c.Redirect(http.StatusFound, "/register/")
	} else if err != nil {
		c.Status(http.StatusInternalServerError)
	}

	c.Redirect(http.StatusFound, "/top/")
}
