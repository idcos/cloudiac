package db

import (
	"cloudiac/configs"
	dbLogger "cloudiac/portal/libs/db/logger"
	"cloudiac/utils/logs"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"os"
	"strconv"
	"strings"
	"time"
)

type Driver struct {
	Dialect    func(dsn string) gorm.Dialector
	SQLEnhance func(sql string) string
	Namer      schema.Namer
}

var (
	drivers = make(map[string]Driver)
)

func GetDriver() (d Driver) {
	dbType := configs.Get().DbType
	if item, ok := drivers[dbType]; ok {
		d = item
	}
	return
}

func Init(dsn string) {
	if err := openDB(dsn); err != nil {
		//logs.Get().Fatalln(err)
		panic(err)
	}
}

func openDB(dsn string, driverNames ...string) error {
	slowThresholdEnv := os.Getenv("GORM_SLOW_THRESHOLD")
	slowThreshold := time.Second
	if slowThresholdEnv != "" {
		n, err := strconv.Atoi(slowThresholdEnv)
		if err != nil {
			return errors.Wrap(err, "GORM_SLOW_THRESHOLD")
		}
		slowThreshold = time.Second * time.Duration(n)
	}

	logLevelEnv := os.Getenv("GORM_LOG_LEVEL")
	logLevel := gormLogger.Warn
	if logLevelEnv != "" {
		switch strings.ToLower(logLevelEnv) {
		case "silent":
			logLevel = gormLogger.Silent
		case "error":
			logLevel = gormLogger.Error
		case "warn", "warning":
			logLevel = gormLogger.Warn
		case "info":
			logLevel = gormLogger.Info
		default:
			logs.Get().Warnf("invalid GORM_LOG_LEVEL '%s'", logLevelEnv)
		}
	}

	driverNameIdx := strings.Index(dsn, "://")
	var driverName string
	if driverNameIdx > 0 {
		driverName, dsn = dsn[0:driverNameIdx], dsn[driverNameIdx+3:]
	} else {
		driverName = "mysql"
	}
	if len(driverNames) > 0 {
		driverName = driverNames[0]
	}

	var dialector gorm.Dialector
	if d, ok := drivers[strings.ToLower(driverName)]; !ok {
		return fmt.Errorf("unsupported db type '%s'", driverName)
	} else {
		dialector = d.Dialect(dsn)
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		NamingStrategy: getNamingStrategy(),
		Logger: dbLogger.New(logs.Get(), gormLogger.Config{
			SlowThreshold:             slowThreshold,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  logLevel,
		}),
	})
	if err != nil {
		return err
	}
	if err = db.Callback().Create().Before("gorm:before_create").
		Register("my_before_create_hook", beforeCreateCallback); err != nil {
		return err
	}
	if err = db.Callback().Create().Before("gorm:before_update").
		Register("my_before_update_hook", beforeUpdateCallback); err != nil {
		return err
	}
	defaultDB = db
	return nil
}
