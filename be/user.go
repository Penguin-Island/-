package be

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	ErrMsgUserName = "ユーザー名は3文字以上で、アルファベットか数字、\"_\"と\"-\"を使うことができます"
	ErrMsgPassword = "パスワードは10文字以上で、アルファベットと数字を含む必要があります"
)

func isValidUserName(userName string) bool {
	if len(userName) < 3 {
		return false
	}

	for _, r := range []rune(userName) {
		if !('a' <= r && r <= 'z') && !('A' <= r && r <= 'Z') && !('0' <= r && r <= '9') && (r != '_') && (r != '-') {
			return false
		}
	}
	return true
}

func isValidPassword(password string) bool {
	if len(password) < 10 {
		return false
	}

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

func registerUser(app *App, userName, password string) (uint, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	member := Member{
		UserName: userName,
		Password: string(hashed),
	}

	if err := app.db.Create(&member).Error; err != nil {
		return 0, err
	}
	return member.ID, nil
}

func handleRegisterUser(app *App, c *gin.Context) {
	userName := c.PostForm("username")
	password := c.PostForm("password")

	if !isValidUserName(userName) {
		c.JSON(http.StatusNotAcceptable, map[string]interface{}{
			"success": false,
			"reason":  ErrMsgUserName,
		})
		return
	}

	if !isValidPassword(password) {
		c.JSON(http.StatusNotAcceptable, map[string]interface{}{
			"success": false,
			"reason":  ErrMsgPassword,
		})
		return
	}

	userId, err := registerUser(app, userName, password)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	sess := sessions.Default(c)
	sess.Set("user_id", userId)
	sess.Save()

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

type GroupInfoResp struct {
	Members    []string `json:"members"`
	WakeUpTime string   `json:"wakeUpTime"`
}

type UserInfoResp struct {
	UserName    string        `json:"userName"`
	JoinedGroup bool          `json:"joinedGroup"`
	GroupInfo   GroupInfoResp `json:"groupInfo"`
	SuccessRate int           `json:"successRate"`
}

func handleGetUserInfo(app *App, c *gin.Context) {
	sess := sessions.Default(c)
	iUserId := sess.Get("user_id")
	if iUserId == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if _, ok := iUserId.(uint); !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	userId := iUserId.(uint)

	var user Member
	if err := app.db.First(&user, userId).Error; err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// 成功率を計算
	daysAfterSignUp, err := getDaysAfterSignUp(app, userId)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	successCount, err := getSuccessCount(app, userId)
	successRate := 100
	if daysAfterSignUp != 0 {
		successRate = successCount * 100 / daysAfterSignUp
	}

	userInfo := UserInfoResp{
		UserName:    user.UserName,
		JoinedGroup: user.GroupId != 0,
		SuccessRate: successRate,
	}

	userInfo.GroupInfo.Members = make([]string, 0)
	if user.GroupId != 0 {
		var group Group
		if err := app.db.First(&group, user.GroupId).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		jst, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		wakeUpTime, err := time.Parse("15:04", group.WakeUpTime)
		if err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		now := time.Now().In(jst)
		wakeUpTime = time.Date(now.Year(), now.Month(), now.Day(), wakeUpTime.Hour(), wakeUpTime.Minute(), 0, 0, time.UTC).In(jst)

		userInfo.GroupInfo.WakeUpTime = fmt.Sprintf("%02d:%02d", wakeUpTime.Hour(), wakeUpTime.Minute())

		var groupMembers []Member
		if err := app.db.Find(&groupMembers, "group_id = ?", user.GroupId).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		for _, memb := range groupMembers {
			if memb.ID != userId {
				userInfo.GroupInfo.Members = append(userInfo.GroupInfo.Members, memb.UserName)
			}
		}
	}

	c.JSON(http.StatusOK, &userInfo)
}

func handleFindUser(app *App, c *gin.Context) {
	sess := sessions.Default(c)
	iUserId := sess.Get("user_id")
	if iUserId == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if _, ok := iUserId.(uint); !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	tag, ok := c.GetQuery("userName")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var user Member
	if err := app.db.First(&user, "user_name = ?", tag).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusFound)
}

type StatisticsResp struct {
	Year    int  `json:"year"`
	Day     int  `json:"day"`
	Month   int  `json:"month"`
	Success bool `json:"success"`
}

func handleGetStatistics(app *App, c *gin.Context) {
	sess := sessions.Default(c)
	iUserId := sess.Get("user_id")
	if iUserId == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if _, ok := iUserId.(uint); !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	userId := iUserId.(uint)

	cacheKey := fmt.Sprintf("stat-%v", userId)

	if result, err := app.redis.Get(context.Background(), cacheKey).Result(); err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == nil {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Writer.WriteString(result)
		return
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var user Member
	if err := app.db.First(&user, userId).Error; err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var statsData []Statistics
	if err := app.db.Where("user_id = ?", user.ID).Order("created_at").Limit(7).Find(&statsData).Error; err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	now := time.Now().In(jst)
	var wakeUpTime time.Time
	if user.GroupId == 0 {
		wakeUpTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jst)
	} else {
		var group Group
		if err := app.db.First(&group, user.GroupId).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		savedTime, err := time.Parse("15:04", group.WakeUpTime)
		if err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		wakeUpTime = time.Date(now.Year(), now.Month(), now.Day(), savedTime.Hour(), savedTime.Minute(), 0, 0, time.UTC).In(jst)
	}

	statistics := collectStats(statsData, wakeUpTime, user.CreatedAt, now, jst)

	jsonData, err := json.Marshal(&statistics)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err := app.redis.Set(context.Background(), cacheKey, string(jsonData), 24*time.Hour).Err(); err != nil {
		log.Error(err)
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Writer.Write(jsonData)
}
