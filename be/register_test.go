package be

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var runnable = true

func init() {
	dbHost := os.Getenv("POSTGRES_HOST")
	if len(dbHost) == 0 {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("POSTGRES_PORT")
	if len(dbPort) == 0 {
		dbHost = "5432"
	}
	dsn := fmt.Sprintf(
		"host=%s user=postgres password= dbname=ohatori port=%s sslmode=disable TimeZone=Asia/Tokyo",
		dbHost, dbPort)
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return
	}
	db = gormDB
}

func Test_registerUser(t *testing.T) {
	if db == nil {
		t.Skip()
	}

	rand.Seed(time.Now().Unix())

	if err := db.Migrator().DropTable(&Member{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrator().CreateTable(&Member{}); err != nil {
		t.Fatal(err)
	}

	app := &App{
		db: db,
	}

	for i := 0; i < 100; i++ {
		acceptable, err := registerUser(app, "hoge", "li]t8OoL")
		if err != nil {
			t.Fatal(err)
		}
		if !acceptable {
			t.Fatal("Unexpected: !acceptable")
		}
	}

	var members []Member
	if err := db.Find(&members).Error; err != nil {
		t.Fatal(err)
	}

	if len(members) != 100 {
		t.Errorf("Not all members inserted: expectedLen=%v, actualLen=%v", 100, len(members))
	}

	if err := db.Migrator().DropTable(&Member{}); err != nil {
		t.Fatal(err)
	}
}
