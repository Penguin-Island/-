package be

import (
	"fmt"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Statistics struct {
	gorm.Model
	UserId  uint
	Success bool
}

func fetchSuccessCount(app *App, userId uint) (int, error) {
	var successCount int64
	if err := app.db.Model(&Statistics{}).Where("user_id = ? AND success = ?", userId, true).Count(&successCount).Error; err != nil {
		return 0, err
	}
	return int(successCount), nil
}

func recordStat(app *App, userId uint, success bool) error {
	if !success {
		return nil
	}

	if err := app.db.Create(&Statistics{UserId: userId, Success: success}).Error; err != nil {
		return err
	}

	// キャッシュを更新
	cacheKey := fmt.Sprintf("nsuccess-%v", userId)
	if _, err := app.memcached.Increment(cacheKey, 1); err != nil && err == memcache.ErrCacheMiss {
		if count, err := fetchSuccessCount(app, userId); err != nil {
			log.Error(err)
		} else {
			if err := app.memcached.Set(&memcache.Item{Key: cacheKey, Value: []byte(strconv.Itoa(count))}); err != nil {
				log.Error(err)
			}
		}
	} else if err != nil {
		log.Error(err)
	}

	return nil
}
