package gormx

import (
	"fmt"
	"os"
	"time"

	"github.com/LouYuanbo1/go-webservice/gormx/config"
	"github.com/LouYuanbo1/go-webservice/gormx/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitGorm(config *config.DBConfig) (*gorm.DB, error) {
	if config == nil {
		return nil, errors.ErrInvalidInitConfig
	}

	var gormDB *gorm.DB
	var dsn string
	var err error

	// 构建时区参数（默认Local）
	timeZone := config.TimeZone
	if timeZone == "" {
		timeZone = "Asia/Shanghai"
	}

	// 构建DSN连接字符串
	switch config.Type {
	case "postgres":
		sslMode := config.Postgres.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			config.Host,
			config.User,
			config.Password,
			config.DBName,
			config.Port,
			sslMode,
			timeZone,
		)
		// 初始化 GORM 数据库连接
		gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(config.LogLevel)), // 设置日志模式为 Info（可选 Silent、Warn、Error）
		})
		if err != nil {
			return nil, errors.NewWithDetails(
				errors.ErrDBConnection,
				"open postgres db",
				config.DBName,
				fmt.Sprintf("host=%s port=%d", config.Host, config.Port),
				err,
			)
		}
	case "mysql":
		tls := config.MySQL.TLS
		if tls == "" {
			tls = "false"
		}
		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=%s&tls=%s",
			config.User,
			config.Password,
			config.Host,
			config.Port,
			config.DBName,
			timeZone,
			tls,
		)
		// 初始化 GORM 数据库连接
		gormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(config.LogLevel)), // 设置日志模式为 Info（可选 Silent、Warn、Error）
		})
		if err != nil {
			return nil, errors.NewWithDetails(
				errors.ErrDBConnection,
				"open mysql db",
				config.DBName,
				fmt.Sprintf("host=%s port=%d", config.Host, config.Port),
				err,
			)
		}
	default:
		return nil, errors.NewWithDetails(
			errors.ErrDBConnection,
			fmt.Sprintf("open %s db", config.Type),
			config.DBName,
			fmt.Sprintf("host=%s port=%d", config.Host, config.Port),
			fmt.Errorf("暂时不支持的数据库类型: %s", config.Type),
		)
	}

	// 读取文件内容
	if config.SchemaFile != "" {
		content, err := os.ReadFile(config.SchemaFile)
		if err != nil {
			return nil, errors.NewWithDetails(
				errors.ErrExecutionSQLScript,
				"read schema file",
				config.DBName,
				fmt.Sprintf("host=%s port=%d", config.Host, config.Port),
				err,
			)
		}

		// 将读取到的内容转换为字符串后执行
		sql := string(content)
		if err := gormDB.Exec(sql).Error; err != nil {
			return nil, errors.NewWithDetails(
				errors.ErrExecutionSQLScript,
				"execute schema file",
				config.DBName,
				fmt.Sprintf("host=%s port=%d", config.Host, config.Port),
				err,
			)
		}
	}

	// 获取底层的 sql.DB 实例以配置连接池
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, errors.NewWithDetails(
			errors.ErrDBConnection,
			"get sql db",
			config.DBName,
			fmt.Sprintf("host=%s port=%d", config.Host, config.Port),
			err,
		)
	}

	// 配置连接池（带默认值逻辑）
	if config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(25) // 默认值
	}

	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(25) // 默认值
	}

	connMaxLifetime, _ := time.ParseDuration(config.ConnMaxLifetime)
	if connMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(connMaxLifetime)
	} else {
		sqlDB.SetConnMaxLifetime(5 * time.Minute) // 默认值
	}

	// 验证连接有效性
	if err := sqlDB.Ping(); err != nil {
		return nil, errors.NewWithDetails(
			errors.ErrDBConnection,
			"ping db",
			config.DBName,
			fmt.Sprintf("host=%s port=%d", config.Host, config.Port),
			err,
		)
	}

	fmt.Println("ConnectGormDB successfully. 成功连接到数据库。")
	return gormDB, nil
}
