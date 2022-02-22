package be

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Statistics struct {
	gorm.Model
	UserId  uint
	Success bool
}

func fetchSuccessCountFromDB(app *App, userId uint) (int, error) {
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
		if count, err := fetchSuccessCountFromDB(app, userId); err != nil {
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

func invalidateStatCache(app *App, userId uint) {
	statCacheKey := fmt.Sprintf("stat-%v", userId)
	if err := app.memcached.Delete(statCacheKey); err != nil {
		log.Error(err)
	}
}

func getSuccessCount(app *App, userId uint) (int, error) {
	cacheKey := fmt.Sprintf("nsuccess-%v", userId)
	if item, err := app.memcached.Get(cacheKey); err != nil && err == memcache.ErrCacheMiss {
		count, err := fetchSuccessCountFromDB(app, userId)
		if err != nil {
			return 0, err
		}

		if err := app.memcached.Set(&memcache.Item{Key: cacheKey, Value: []byte(strconv.Itoa(count))}); err != nil {
			log.Error(err)
		}

		return count, nil
	} else if err != nil {
		return 0, err
	} else {
		return strconv.Atoi(string(item.Value))
	}
}

func durationDays(from, to time.Time) int {
	diff := to.Sub(from)
	return int(diff.Hours()) / 24
}

func getDaysAfterSignUp(app *App, userId uint) (int, error) {
	var user Member
	if err := app.db.First(&user, userId).Error; err != nil {
		return 0, err
	}

	registrationDate := user.CreatedAt
	now := time.Now()
	return durationDays(registrationDate, now), nil
}

func collectStats(stats []Statistics, wakeUpTime, signUpTime, until time.Time, tz *time.Location) []StatisticsResp {
	if len(stats) == 0 {
		return make([]StatisticsResp, 0)
	}

	until = until.In(tz)
	signUpTime = signUpTime.In(tz)
	wakeUpTime = wakeUpTime.In(tz)

	if until.Before(wakeUpTime) {
		until.Add(-24 * time.Hour)
	}

	until = time.Date(until.Year(), until.Month(), until.Day(), signUpTime.Hour(), signUpTime.Minute(), signUpTime.Second(), signUpTime.Nanosecond(), tz)

	duration := durationDays(signUpTime, until)
	if duration > 7 {
		duration = 7
	} else {
		firstWakeUp := time.Date(signUpTime.Year(), signUpTime.Month(), signUpTime.Day(), wakeUpTime.Hour(), wakeUpTime.Minute(), 0, 0, tz)
		if firstWakeUp.After(signUpTime) {
			duration += 1
		}
	}
	result := make([]StatisticsResp, duration)
	dataInd := len(stats) - 1
	for i := 0; i < duration; i++ {
		day := until.Add(-time.Duration(24*i) * time.Hour)

		for j := dataInd; j >= 0; j-- {
			dataDay := stats[j].CreatedAt.In(tz)
			if dataDay.Year() == day.Year() && dataDay.Month() == day.Month() && dataDay.Day() == day.Day() {
				dataInd = j - 1
				result[i].Success = stats[j].Success
				break
			}
			if dataDay.Before(day) {
				dataInd = j
				break
			}
		}

		result[i].Year = day.Year()
		result[i].Month = int(day.Month())
		result[i].Day = day.Day()
	}

	return result
}
