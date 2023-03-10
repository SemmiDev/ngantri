package utils

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"ngantri/pawnshop"
	"ngantri/queue"
	"time"
)

func ConnectDB() (*gorm.DB, error) {
	dsn := GetEnv("DSN", "root:@tcp(127.0.0.1:3306)/ngantri?charset=utf8mb4&parseTime=True&loc=Local")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&queue.Queue{},
		&pawnshop.Pawnshop{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
