package tests_test

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	iris "github.com/caretdev/gorm-iris"
	iriscontainer "github.com/caretdev/testcontainers-iris-go"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	. "gorm.io/gorm/utils/tests"
)

var DB *gorm.DB

var connectionString string = "iris://_SYSTEM:SYS@localhost:1972/USER"

var container *iriscontainer.IRISContainer = nil

func TestMain(m *testing.M) {
	var (
		useContainer   bool
		containerImage string
	)
	flag.BoolVar(&useContainer, "container", true, "Use container image.")
	flag.StringVar(&containerImage, "container-image", "", "Container image.")
	flag.Parse()
	var err error
	ctx := context.Background()
	if useContainer || containerImage != "" {
		if containerImage != "" {
			container, err = iriscontainer.Run(ctx, containerImage)
		} else {
			container, err = iriscontainer.RunContainer(ctx)
		}
		if err != nil {
			log.Println("Failed to start container:", err)
			os.Exit(1)
		}
		defer container.Terminate(ctx)
		connectionString = container.MustConnectionString(ctx)
		log.Println("Container started successfully", connectionString)
	}

	var exitCode int = 0
	if DB, err = OpenTestConnection(&gorm.Config{}); err != nil {
		log.Printf("failed to connect database, got error %v", err)
		os.Exit(1)
	} else {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Printf("failed to connect database, got error %v", err)
			os.Exit(1)
		}
		defer sqlDB.Close()

		err = sqlDB.Ping()
		if err != nil {
			log.Printf("failed to ping sqlDB, got error %v", err)
			os.Exit(1)
		}

		RunMigrations()

		exitCode = m.Run()
	}
	if container != nil {
		container.Terminate(ctx)
	}
	os.Exit(exitCode)
}

func OpenTestConnection(cfg *gorm.Config) (db *gorm.DB, err error) {
	// return gorm.Open(sqlite.Open("::memory:"), &gorm.Config{})
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Adjust this threshold (e.g., 1 second)
			LogLevel:                  logger.Warn, // Log level: Silent, Error, Warn, Info
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound
			Colorful:                  true,        // Enable colorized output
		},
	)
	dbDSN := connectionString
	cfg.Logger = newLogger
	db, err = gorm.Open(iris.New(iris.Config{
		DSN: dbDSN,
	}), cfg)
	return
}

func RunMigrations() {
	var err error
	allModels := []interface{}{&User{}, &Account{}, &Pet{}, &Company{}, &Toy{}, &Language{}, &Coupon{}, &CouponProduct{}, &Order{}, &Parent{}, &Child{}, &Tools{}}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allModels), func(i, j int) { allModels[i], allModels[j] = allModels[j], allModels[i] })

	DB.Migrator().DropTable("user_friends", "user_speaks")

	if err = DB.Migrator().DropTable(allModels...); err != nil {
		log.Printf("Failed to drop table, got error %v\n", err)
		os.Exit(1)
	}

	if err = DB.AutoMigrate(allModels...); err != nil {
		log.Printf("Failed to auto migrate, but got error %v\n", err)
		os.Exit(1)
	}

	for _, m := range allModels {
		if !DB.Migrator().HasTable(m) {
			log.Printf("Failed to create table for %#v\n", m)
			os.Exit(1)
		}
	}
}
