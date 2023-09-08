// Copyright © 2023 zc2638 <zc2638@qq.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package database

import (
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	Debug       bool   `json:"debug"`
	Driver      string `json:"driver"`
	DSN         string `json:"dsn"`
	MaxOpenConn int    `json:"maxOpenConn"`
}

func (c *Config) Clone() *Config {
	return &Config{
		Debug:       c.Debug,
		Driver:      c.Driver,
		DSN:         c.DSN,
		MaxOpenConn: c.MaxOpenConn,
	}
}

func New(cfg Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch strings.ToLower(cfg.Driver) {
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	case "mysql":
		conf, err := mysqlDriver.ParseDSN(cfg.DSN)
		if err != nil {
			return nil, err
		}
		conf.ParseTime = true
		dialector = mysql.New(mysql.Config{DSNConfig: conf})
	case "sqlite", "sqlite3":
		dialector = sqlite.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported driver: %v", cfg.Driver)
	}

	config := &gorm.Config{}
	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, err
	}
	if cfg.Debug {
		db = db.Debug()
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)
	return db, nil
}

func AutoDatabase(cfg Config) error {
	var (
		initSQL string
		dsn     string
	)
	switch strings.ToLower(cfg.Driver) {
	case "postgres":
		return nil
	case "mysql":
		conf, err := mysqlDriver.ParseDSN(cfg.DSN)
		if err != nil {
			return err
		}
		initSQL = "CREATE DATABASE IF NOT EXISTS `" + conf.DBName + "` DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci"
		conf.DBName = ""
		dsn = conf.FormatDSN()
	case "sqlite", "sqlite3":
		return nil
	default:
		return fmt.Errorf("unsupported driver: %v", cfg.Driver)
	}

	db, err := New(Config{Debug: true, Driver: cfg.Driver, DSN: dsn})
	if err != nil {
		return err
	}
	if err := db.Exec(initSQL).Error; err != nil {
		return fmt.Errorf("create database failed: %v", err)
	}
	return nil
}
