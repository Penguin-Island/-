package be

import (
	"errors"
	"net/http"
	"strconv"

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

type InvitationResp struct {
	Id      uint   `json:"invitationId"`
	Invitee string `json:"invitee"`
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
		var invitee Member
		if err := app.db.First(&invitee, inv.Invitee).Error; err != nil {
			log.Error(err)
			continue
		}

		invitationResp = append(invitationResp, InvitationResp{
			Id:      inv.ID,
			Invitee: invitee.PlayerTag,
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

		var count int64
		if err := tx.Model(&Member{}).Where("group_id = ?", invitation.GroupId).Count(&count).Error; err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return err
		}
		if count > 1 {
			c.AbortWithStatus(http.StatusNotAcceptable)
			return errors.New("too many members")
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

}
