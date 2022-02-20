package be

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
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

	if acceptable, err := registerUser(app, userName, password); !acceptable {
		c.Redirect(http.StatusFound, "/register/")
	} else if err != nil {
		c.Status(http.StatusInternalServerError)
	}

	c.Redirect(http.StatusFound, "/")
}

type GroupInfoResp struct {
	Members    []string `json:"members"`
	WakeUpTime string   `json:"wakeUpTime"`
}

type UserInfoResp struct {
	UserName    string        `json:"userName"`
	PlayerTag   string        `json:"playerTag"`
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

	userInfo := UserInfoResp{
		UserName:    user.UserName,
		PlayerTag:   user.PlayerTag,
		JoinedGroup: user.GroupId != 0,
		SuccessRate: 100,
	}

	userInfo.GroupInfo.Members = make([]string, 0)
	if user.GroupId != 0 {
		var group Group
		if err := app.db.First(&group, user.GroupId).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		userInfo.GroupInfo.WakeUpTime = group.WakeUpTime

		var groupMembers []Member
		if err := app.db.Find(&groupMembers, "group_id = ?", user.GroupId).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		for _, memb := range groupMembers {
			if memb.ID != userId {
				userInfo.GroupInfo.Members = append(userInfo.GroupInfo.Members, memb.PlayerTag)
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

	tag, ok := c.GetQuery("playerTag")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	var user Member
	if err := app.db.First(&user, "player_tag = ?", tag).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusFound)
}
