package be

import (
	"time"

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

func fetchFailureCountFromDB(app *App, userId uint) (int, error) {
	var failureCount int64
	if err := app.db.Model(&Statistics{}).Where("user_id = ? AND success = ?", userId, false).Count(&failureCount).Error; err != nil {
		return 0, err
	}
	return int(failureCount), nil
}

func recordStat(app *App, userId uint, success bool) error {
	if err := app.db.Create(&Statistics{UserId: userId, Success: success}).Error; err != nil {
		return err
	}

	return nil
}

func getSuccessCount(app *App, userId uint) (int, error) {
	count, err := fetchSuccessCountFromDB(app, userId)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func getFailureCount(app *App, userId uint) (int, error) {
	count, err := fetchFailureCountFromDB(app, userId)
	if err != nil {
		return 0, err
	}

	return count, nil
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
	return durationDays(registrationDate, now) + 1, nil
}

func collectStats(stats []Statistics, wakeUpTime, signUpTime, until time.Time, tz *time.Location) []StatisticsResp {
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
