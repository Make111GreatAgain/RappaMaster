package Database

import (
	"BHLayer2Node/Date"
	"BHLayer2Node/paradigm"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseService struct {
	db      *gorm.DB
	dbConf  *paradigm.DatabaseConfig
	channel *paradigm.RappaChannel
}

// 全局单例
//var (
//	dbService     *DatabaseService
//	dbServiceOnce sync.Once
//)

func NewDatabaseService(config *paradigm.BHLayer2NodeConfig, channel *paradigm.RappaChannel) (*DatabaseService, error) {
	var initErr error
	// 构建DSN
	dsn := config.Database.BuildDSN()

	// 初始化GORM
	gormConfig := gorm.Config{}
	if config.DEBUG {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}
	gormDB, err := gorm.Open(mysql.Open(dsn), &gormConfig)
	if err != nil {
		initErr = fmt.Errorf("failed to open database: %v", err)
		return nil, err
	}

	// 配置连接池
	sqlDB, err := gormDB.DB()
	if err != nil {
		initErr = fmt.Errorf("failed to get sql.DB: %v", err)
		return nil, err
	}
	sqlDB.SetMaxIdleConns(config.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Database.MaxOpenConns)
	maxLifetime, _ := time.ParseDuration(config.Database.MaxLifetime)
	sqlDB.SetConnMaxLifetime(maxLifetime)

	return &DatabaseService{
		db:      gormDB,
		dbConf:  config.Database,
		channel: channel,
	}, initErr
}

func (o DatabaseService) AutoMigrate() error {
	return o.db.AutoMigrate(
		&paradigm.PlatformTask{},
		&paradigm.Slot{},
		&paradigm.Task{},
		&paradigm.DevEpoch{},
		&paradigm.DevReference{},
		&Date.DateRecord{},
	)
}

func (o DatabaseService) TruncateAll() error {
	tables := []string{"date_records", "dev_epoches", "dev_references", "slots", "tasks", "platform_tasks"}
	return o.db.Connection(func(tx *gorm.DB) (err error) {
		if err = tx.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
			return fmt.Errorf("failed to disable foreign key checks: %w", err)
		}

		defer func() {
			if enableErr := tx.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; enableErr != nil && err == nil {
				err = fmt.Errorf("failed to enable foreign key checks: %w", enableErr)
			}
		}()

		for _, table := range tables {
			if err = tx.Exec(fmt.Sprintf("TRUNCATE TABLE `%s`", table)).Error; err != nil {
				return fmt.Errorf("failed to truncate table %s: %w", table, err)
			}
		}
		return nil
	})
}
