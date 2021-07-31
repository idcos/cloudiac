package db

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"cloudiac/portal/consts/e"
	"cloudiac/utils/logs"
)

var defaultSess *gorm.DB
var logger logs.Logger

type Session struct {
	db *gorm.DB
}

func (s *Session) Begin() *Session {
	return ToSess(s.db.Begin())
}

func (s *Session) GormDB() *gorm.DB {
	return s.db
}

func (s *Session) New() *Session {
	return ToSess(s.db.Session(&gorm.Session{}))
}

func (s *Session) AddUniqueIndex(indexName string, columns ...string) error {
	return s.db.AddUniqueIndex(indexName, columns...).Error
}

func (s *Session) RemoveIndex(table string, indexName string) error {
	if !s.db.Dialect().HasIndex(table, indexName) {
		return nil
	}

	return s.db.Table(table).RemoveIndex(indexName).Error
}

func (s *Session) DropColumn(table string, columns ...string) error {
	migrator := s.db.Migrator()
	for _, col := range columns {
		if migrator.HasColumn(table, col) {
			if err := migrator.DropColumn(table, col); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Session) ModifyColumn(column string, typ string) error {
	return s.db.ModifyColumn(column, typ).Error
}

func (s *Session) Rollback() error {
	return s.db.Rollback().Error
}

func (s *Session) Commit() error {
	return s.db.Commit().Error
}

func (s *Session) Model(m interface{}) *Session {
	return ToSess(s.db.Model(m))
}

func (s *Session) Table(name string) *Session {
	return ToSess(s.db.Table(name))
}

func (s *Session) Debug() *Session {
	return ToSess(s.db.Debug())
}

func (s *Session) Expr() interface{} {
	return s.db
}

func (s *Session) Raw(sql string, values ...interface{}) *Session {
	return ToSess(s.db.Raw(sql, values...))
}

func (s *Session) Exec(sql string, args ...interface{}) (int64, error) {
	r := s.db.Exec(sql, args...)
	return r.RowsAffected, r.Error
}

func (s *Session) Rows() (*sql.Rows, error) {
	return s.db.Rows()
}

func (s *Session) IterRows(dest interface{}, fn func() error) error {
	rows, err := s.db.Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(dest); err != nil {
			return err
		}
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Session) Row() *sql.Row {
	return s.db.Row()
}

func (s *Session) Unscoped() *Session {
	return ToSess(s.db.Unscoped())
}

func (s *Session) Select(query interface{}, args ...interface{}) *Session {
	return ToSess(s.db.Select(query, args...))
}

func (s *Session) LazySelect(selectStat ...string) *Session {
	return ToSess(s.db.Set("app:lazySelects", selectStat))
}

func (s *Session) LazySelectAppend(selectStat ...string) *Session {
	stats, ok := s.db.Get("app:lazySelects")
	if ok {
		return ToSess(s.db.Set("app:lazySelects", append(stats.([]string), selectStat...)))
	} else {
		return ToSess(s.db.Set("app:lazySelects", selectStat))
	}
}

func (s *Session) Where(query interface{}, args ...interface{}) *Session {
	return ToSess(s.db.Where(query, args...))
}

func (s *Session) WhereLike(col string, pattern string) *Session {
	return ToSess(s.db.Where("? LIKE ?", gorm.Expr(col), "%"+pattern+"%"))
}

func (s *Session) Joins(query string, args ...interface{}) *Session {
	return ToSess(s.db.Joins(query, args...))
}

func (s *Session) Order(value interface{}) *Session {
	return ToSess(s.db.Order(value))
}

func (s *Session) Group(query string) *Session {
	return ToSess(s.db.Group(query))
}

func (s *Session) Limit(n int) *Session {
	return ToSess(s.db.Limit(n))
}

func (s *Session) Offset(n int) *Session {
	return ToSess(s.db.Offset(n))
}

func (s *Session) Set(name string, value interface{}) *Session {
	return ToSess(s.db.Set(name, value))
}

func (s *Session) Count() (cnt int64, err error) {
	err = s.db.Count(&cnt).Error
	return
}

func (s *Session) Exists() (bool, error) {
	exists := false
	err := s.New().Raw("SELECT EXISTS(?)", s.Expr()).Row().Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Session) autoLazySelect() *Session {
	selects, ok := s.db.Get("app:lazySelects")
	if !ok {
		return s
	}
	return s.Select(strings.Join(selects.([]string), ","))
}

func (s *Session) First(out interface{}) error {
	qs := s.autoLazySelect()
	return qs.db.First(out).Error
}

func (s *Session) Last(out interface{}) error {
	qs := s.autoLazySelect()
	return qs.db.Last(out).Error
}

func (s *Session) Find(out interface{}, where ...interface{}) error {
	qs := s.autoLazySelect()
	err := qs.db.Find(out, where...).Error
	if e.IsRecordNotFound(err) {
		return nil
	} else {
		return err
	}
}

func (s *Session) Scan(out interface{}) error {
	qs := s.autoLazySelect()
	err := qs.db.Scan(out).Error
	return err
}

func (s *Session) Delete(val interface{}) (int64, error) {
	r := s.db.Delete(val)
	return r.RowsAffected, r.Error
}

// Update
// example: Update("name", "newName") or Update(map[string]interface{}{"name":"newName"})
func (s *Session) Update(attrs ...interface{}) (int64, error) {
	var r *gorm.DB
	if len(attrs) == 2 {
		r = s.db.Update(attrs[0].(string), attrs[1])
	} else {
		r = s.db.Updates(attrs[0])
	}
	return r.RowsAffected, r.Error
}

func (s *Session) Save(val interface{}) (int64, error) {
	r := s.db.Save(val)
	return r.RowsAffected, r.Error
}

func (s *Session) Insert(val interface{}) error {
	return s.db.Create(val).Error
}

func (s *Session) Pluck(column string, val interface{}) error {
	return s.db.Pluck(column, val).Error
}

/*
CompareFieldValue 将指定的 field 与 q 比较大小
q 的格式: 10, gt:10, ge:10, lt:10, le:10
*/
func (s *Session) CompareFieldValue(field string, q string) (*Session, error) {
	var (
		invalidErr = fmt.Errorf("invalid compare query: %s", q)
	)

	if pos := strings.IndexByte(q, ':'); pos > 0 {
		op := q[0:pos]
		val, err := strconv.Atoi(q[pos+1:])
		if err != nil {
			return nil, invalidErr
		}

		switch op {
		case "gt":
			return s.Where(fmt.Sprintf("%s > ?", field), val), nil
		case "ge":
			return s.Where(fmt.Sprintf("%s >= ?", field), val), nil
		case "lt":
			return s.Where(fmt.Sprintf("%s < ?", field), val), nil
		case "le":
			return s.Where(fmt.Sprintf("%s <= ?", field), val), nil
		default:
			return nil, invalidErr
		}
	} else {
		val, err := strconv.Atoi(q)
		if err != nil {
			return nil, invalidErr
		}
		return s.Where(fmt.Sprintf("%s = ?", field), val), nil
	}
}

func ToSess(db *gorm.DB) *Session {
	return &Session{db: db}
}

func ToColName(name string) string {
	return gorm.ToColumnName(name)
}

func Get() *Session {
	return ToSess(defaultSess)
}

func openDB(dsn string) error {
	db, err := gorm.Open(mysql.Open(dsn), nil)
	if err != nil {
		return err
	}

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

	defaultSess = db.Session(&gorm.Session{
		Logger: gormLogger.New(logs.Get(), gormLogger.Config{
			SlowThreshold:             slowThreshold,
			Colorful:                  false,
			IgnoreRecordNotFoundError: false,
			LogLevel:                  logLevel,
		}),
	})
	return nil
}

func Init(dsn string) {
	logger = logs.Get()
	if err := openDB(dsn); err != nil {
		logger.Fatalln(err)
	}
	//db.SetLogger(sqlLogger{logger})
}
