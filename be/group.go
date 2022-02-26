package be

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Group struct {
	gorm.Model
	WakeUpTime string `gorm:"default:'22:00'"`
}

type Invitation struct {
	gorm.Model
	Inviter uint
	Invitee uint
	GroupId uint
}

type InvitationResp struct {
	Id      uint   `json:"invitationId"`
	Inviter string `json:"inviter"`
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
		if err := tx.First(&invitee, "user_name = ?", inviteeTag).Error; err != nil {
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

	var invitations []Invitation
	if err := app.db.Find(&invitations, "invitee = ?", userId).Error; err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	invitationResp := make([]InvitationResp, 0)
	for _, inv := range invitations {
		var inviter Member
		if err := app.db.First(&inviter, inv.Inviter).Error; err != nil {
			log.Error(err)
			continue
		}

		invitationResp = append(invitationResp, InvitationResp{
			Id:      inv.ID,
			Inviter: inviter.UserName,
		})
	}

	c.JSON(http.StatusOK, invitationResp)
}

func handleJoin(app *App, c *gin.Context) {
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
	sInvitationId, ok := c.GetPostForm("invitationId")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	invitationId, err := strconv.ParseUint(sInvitationId, 10, 64)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
	}

	app.db.Transaction(func(tx *gorm.DB) error {
		var user Member
		if err := tx.First(&user, userId).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return err
		}

		var invitation Invitation
		if err := tx.First(&invitation).Where("id = ?", invitationId, "invitee = ?", userId).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusNotAcceptable)
			return err
		}

		user.GroupId = invitation.GroupId
		if err := tx.Save(user).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return err
		}

		if err := tx.Delete(&invitation).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return err
		}

		c.Status(http.StatusAccepted)

		return nil
	})
}

func handleUnjoin(app *App, c *gin.Context) {
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

	if err := app.db.Model(&Member{}).Where(userId).Update("group_id", 0).Error; err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

func handleDeclineInvitations(app *App, c *gin.Context) {
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
	sInvitationId, ok := c.GetPostForm("invitationId")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	invitationId, err := strconv.ParseUint(sInvitationId, 10, 64)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
	}

	if err := app.db.Model(&Invitation{}).Where("invitee = ?", userId).Delete("id = ?", invitationId).Error; err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

func handleSetTime(app *App, c *gin.Context) {
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

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	parsedTime, err := time.Parse("15:04", c.PostForm("time"))
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	now := time.Now().In(jst)
	resultTime := time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, jst).In(time.UTC)

	var user Member
	if err := app.db.First(&user, userId).Error; err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err := app.db.Model(&Group{}).Where(user.GroupId).Update("wake_up_time", fmt.Sprintf("%02d:%02d", resultTime.Hour(), resultTime.Minute())).Error; err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}
