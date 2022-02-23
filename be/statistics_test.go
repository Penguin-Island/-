package be

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

func Test_durationDays(t *testing.T) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}

	testcases := []struct {
		name   string
		from   time.Time
		to     time.Time
		result int
	}{
		{
			name:   "just",
			from:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			to:     time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC),
			result: 1,
		},
		{
			name:   "different time",
			from:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			to:     time.Date(2000, 1, 3, 10, 5, 0, 0, time.UTC),
			result: 2,
		},
		{
			name:   "different locale",
			from:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			to:     time.Date(2000, 1, 4, 8, 5, 0, 0, jst),
			result: 2,
		},
		{
			name:   "different month",
			from:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			to:     time.Date(2000, 2, 4, 8, 5, 0, 0, time.UTC),
			result: 34,
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			result := durationDays(testcase.from, testcase.to)
			if result != testcase.result {
				if result != testcase.result {
					t.Errorf("Unexpected result for [%v-%v]: expected=%v, actual=%v\n", testcase.from, testcase.to, testcase.result, result)
				}
			}
		})
	}
}

func Test_collectStats(t *testing.T) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}

	testcases := []struct {
		name       string
		stats      []Statistics
		wakeUpTime time.Time
		signUpTime time.Time
		until      time.Time
		result     []StatisticsResp
	}{
		{
			name: "simple",
			stats: []Statistics{
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 10, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 11, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: false,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 12, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 13, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 14, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 15, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 16, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
			},
			wakeUpTime: time.Date(0, 1, 1, 7, 15, 0, 0, time.UTC),
			signUpTime: time.Date(2022, 2, 1, 10, 5, 5, 0, jst).In(time.UTC),
			until:      time.Date(2022, 2, 16, 7, 21, 0, 0, jst),
			result: []StatisticsResp{
				{
					Year:    2022,
					Month:   2,
					Day:     16,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     15,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     14,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     13,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     12,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     11,
					Success: false,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     10,
					Success: true,
				},
			},
		},
		{
			name: "duplicate",
			stats: []Statistics{
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 10, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 11, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: false,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 11, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: false,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 12, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 13, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 14, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 15, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 16, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
			},
			wakeUpTime: time.Date(2022, 2, 16, 7, 15, 0, 0, jst).In(time.UTC),
			signUpTime: time.Date(2022, 2, 1, 10, 5, 5, 0, jst).In(time.UTC),
			until:      time.Date(2022, 2, 16, 7, 21, 0, 0, jst),
			result: []StatisticsResp{
				{
					Year:    2022,
					Month:   2,
					Day:     16,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     15,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     14,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     13,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     12,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     11,
					Success: false,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     10,
					Success: true,
				},
			},
		},
		{
			name: "missing",
			stats: []Statistics{
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 10, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 11, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: false,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 13, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 14, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 16, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
			},
			wakeUpTime: time.Date(2022, 2, 16, 7, 15, 0, 0, jst).In(time.UTC),
			signUpTime: time.Date(2022, 2, 1, 10, 5, 5, 0, jst).In(time.UTC),
			until:      time.Date(2022, 2, 16, 7, 21, 0, 0, jst),
			result: []StatisticsResp{
				{
					Year:    2022,
					Month:   2,
					Day:     16,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     15,
					Success: false,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     14,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     13,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     12,
					Success: false,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     11,
					Success: false,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     10,
					Success: true,
				},
			},
		},
		{
			name: "immature account 1",
			stats: []Statistics{
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 13, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 14, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 16, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
			},
			wakeUpTime: time.Date(2022, 2, 16, 7, 15, 0, 0, jst).In(time.UTC),
			signUpTime: time.Date(2022, 2, 12, 7, 18, 5, 0, jst).In(time.UTC),
			until:      time.Date(2022, 2, 16, 7, 21, 0, 0, jst),
			result: []StatisticsResp{
				{
					Year:    2022,
					Month:   2,
					Day:     16,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     15,
					Success: false,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     14,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     13,
					Success: true,
				},
			},
		},
		{
			name: "immature account 2",
			stats: []Statistics{
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 13, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 14, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 16, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
			},
			wakeUpTime: time.Date(2022, 2, 16, 7, 15, 0, 0, jst).In(time.UTC),
			signUpTime: time.Date(2022, 2, 12, 10, 55, 5, 0, jst).In(time.UTC),
			until:      time.Date(2022, 2, 16, 7, 21, 0, 0, jst),
			result: []StatisticsResp{
				{
					Year:    2022,
					Month:   2,
					Day:     16,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     15,
					Success: false,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     14,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     13,
					Success: true,
				},
			},
		},
		{
			name: "immature account 3",
			stats: []Statistics{
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 12, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 13, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 14, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 16, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
			},
			wakeUpTime: time.Date(2022, 2, 16, 7, 15, 0, 0, jst).In(time.UTC),
			signUpTime: time.Date(2022, 2, 12, 5, 55, 5, 0, jst).In(time.UTC),
			until:      time.Date(2022, 2, 16, 7, 21, 0, 0, jst),
			result: []StatisticsResp{
				{
					Year:    2022,
					Month:   2,
					Day:     16,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     15,
					Success: false,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     14,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     13,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     12,
					Success: true,
				},
			},
		},
		{
			name: "immature account 4",
			stats: []Statistics{
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 12, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 13, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 14, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 16, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
			},
			wakeUpTime: time.Date(2022, 2, 16, 7, 15, 0, 0, jst).In(time.UTC),
			signUpTime: time.Date(2022, 2, 12, 10, 55, 5, 0, jst).In(time.UTC),
			until:      time.Date(2022, 2, 16, 7, 0, 0, 0, jst),
			result: []StatisticsResp{
				{
					Year:    2022,
					Month:   2,
					Day:     16,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     15,
					Success: false,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     14,
					Success: true,
				},
				{
					Year:    2022,
					Month:   2,
					Day:     13,
					Success: true,
				},
			},
		},
		{
			name: "immature account 5",
			stats: []Statistics{
				{
					Model: gorm.Model{
						CreatedAt: time.Date(2022, 2, 16, 7, 20, 10, 0, jst).In(time.UTC),
					},
					Success: true,
				},
			},
			wakeUpTime: time.Date(2022, 2, 16, 7, 15, 0, 0, jst).In(time.UTC),
			signUpTime: time.Date(2022, 2, 16, 7, 10, 5, 0, jst).In(time.UTC),
			until:      time.Date(2022, 2, 16, 7, 23, 0, 0, jst),
			result: []StatisticsResp{
				{
					Year:    2022,
					Month:   2,
					Day:     16,
					Success: true,
				},
			},
		},
		{
			name:       "immature account 6",
			stats:      []Statistics{},
			wakeUpTime: time.Date(2022, 2, 16, 7, 15, 0, 0, jst).In(time.UTC),
			signUpTime: time.Date(2022, 2, 16, 7, 10, 5, 0, jst).In(time.UTC),
			until:      time.Date(2022, 2, 16, 7, 23, 0, 0, jst),
			result: []StatisticsResp{
				{
					Year:    2022,
					Month:   2,
					Day:     16,
					Success: false,
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			result := collectStats(testcase.stats, testcase.wakeUpTime, testcase.signUpTime, testcase.until, jst)

			if len(result) != len(testcase.result) {
				t.Fatalf("Unexpected length:\n\texpects=%v\n\tactual=%v", len(testcase.result), len(result))
			}

			for i, resp := range result {
				if testcase.result[i].Year != resp.Year {
					t.Fatalf("Unexpected year at %v:\n\texpects=%v\n\tactual=%v", i, testcase.result[i].Year, resp.Year)
				}
				if testcase.result[i].Month != resp.Month {
					t.Fatalf("Unexpected month at %v:\n\texpects=%v\n\tactual=%v", i, testcase.result[i].Month, resp.Month)
				}
				if testcase.result[i].Day != resp.Day {
					t.Fatalf("Unexpected day at %v:\n\texpects=%v\n\tactual=%v", i, testcase.result[i].Day, resp.Day)
				}
				if testcase.result[i].Success != resp.Success {
					t.Fatalf("Unexpected result at %v:\n\texpects=%v\n\tactual=%v", i, testcase.result[i].Success, resp.Success)
				}
			}
		})
	}
}
