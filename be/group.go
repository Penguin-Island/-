package be

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Group struct {
	gorm.Model
}

type Invitation struct {
	gorm.Model
	Inviter uint
	Invitee uint
	GroupId uint
}

func handleInvite(app *App, c *gin.Context) {
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
	inviteeTag, ok := c.GetPostForm("player")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	app.db.Transaction(func(tx *gorm.DB) error {
		var memb Member
		if err := tx.First(&memb, userId).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return err
		}

		var invitee Member
		if err := tx.First(&invitee, "player_tag = ?", inviteeTag).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return err
		}

		groupId := memb.GroupId
		if memb.GroupId == 0 {
			// ユーザが何のグループにも所属していないときは新しいグループを作成する

			group := Group{}
			if err := tx.Create(&group).Error; err != nil {
				log.Error(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return err
			}
			memb.GroupId = group.ID
			if err := tx.Save(&memb).Error; err != nil {
				log.Error(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return err
			}

			groupId = group.ID
		}

		invitation := Invitation{
			Inviter: uint(userId),
			Invitee: invitee.ID,
			GroupId: groupId,
		}
		if err := tx.Create(&invitation).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return err
		}

		c.Status(http.StatusCreated)

		return nil
	})
}

func handleGetInvitations(app *App, c *gin.Context) {

}

func handleJoin(app *App, c *gin.Context) {

}

func handleUnjoin(app *App, c *gin.Context) {

}
