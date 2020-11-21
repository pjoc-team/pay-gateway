package db

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/pjoc-team/tracing/logger"
	"regexp"
	"strings"
	"time"
)

// MysqlConfig mysql配置
type MysqlConfig struct {
	URL     string        `yaml:"url" json:"url"`
	MaxConn int           `yaml:"max_conn" json:"max_conn"`
	MaxIdle time.Duration `yaml:"max_idle" json:"max_idle"`
}

// InitDb 连接db
func InitDb(ctx context.Context, config *MysqlConfig) (*gorm.DB, error) {
	log := logger.ContextLog(ctx)
	// db, err := gorm.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=True&loc=Local")
	log.Infof("connecting mysql: %v", GetHost(config.URL))
	db, err := gorm.Open("mysql", config.URL)
	if err != nil {
		log.Errorf("failed to init db: %v error: %v", config, err.Error())
		return nil, err
	}
	log.Infof("connected db: %v", config.URL)
	db.DB().SetConnMaxLifetime(time.Duration(config.MaxIdle) * time.Second)
	db.DB().SetMaxOpenConns(config.MaxConn)

	if log.IsDebugEnabled() {
		db = db.Debug()
	}

	return db, nil
}

// GetHost 获取db配置里面的host
func GetHost(url string) string {
	if url == "" {
		return ""
	}
	regex := regexp.MustCompile(`\(.*?\)`)
	findString := regex.FindString(url)
	findString = strings.TrimPrefix(findString, "(")
	findString = strings.TrimSuffix(findString, ")")
	return findString
}
