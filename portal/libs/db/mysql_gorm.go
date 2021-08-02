package db

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm/schema"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/plugin/soft_delete"

	"cloudiac/portal/consts/e"
	"cloudiac/utils/logs"
)

var (
	defaultDB      *gorm.DB
	namingStrategy = schema.NamingStrategy{}
)

type SoftDeletedAt soft_delete.DeletedAt

type Session struct {
	db *gorm.DB
}

func (s *Session) Begin() *Session {
	return ToSess(s.db.Begin())
}

func (s *Session) Transaction(fc func(tx *Session) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return fc(ToSess(tx))
	})
}

func (s *Session) GormDB() *gorm.DB {
	return s.db
}

func (s *Session) New() *Session {
	return ToSess(s.db.Session(&gorm.Session{NewDB: true}))
}

func (s *Session) AddUniqueIndex(indexName string, columns ...string) error {
	stmt := s.db.Statement
	if stmt.Model != nil {
		if err := stmt.Parse(stmt.Model); err != nil {
			return err
		}
	}

	if s.db.Migrator().HasIndex(stmt.Table, indexName) {
		return nil
	}

	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}

	for i := range columns {
		columns[i] = fmt.Sprintf("`%s`", columns[i])
	}
	_, err = sqlDB.Exec(fmt.Sprintf("CREATE UNIQUE INDEX `%s` ON %s (%s)",
		indexName, stmt.Table, strings.Join(columns, ",")))
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) RemoveIndex(table string, indexName string) error {
	migrator := s.db.Migrator()
	if migrator.HasIndex(table, indexName) {
		return migrator.DropIndex(table, indexName)
	}
	return nil
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

func (s *Session) ModifyColumn(table interface{}, column string) error {
	//logs.Get().Debugf("s.db.Statement: %#v", s.db.Statement)
	//logs.Get().Debugf("s.db.Statement.Schema: %#v", s.db.Statement.Schema)
	////logs.Get().Debugf("s.db.Statement.Schema.LookUpField: %#v", s.db.Statement.Schema.LookUpField)
	//return s.db.Migrator().AlterColumn(table, column)
	// TODO fixme
	return nil
}

func (s *Session) Rollback() error {
	return s.db.Rollback().Error
}

func (s *Session) Commit() error {
	return s.db.Commit().Error
}

func (s *Session) Model(m interface{}) *Session {
	switch v := m.(type) {
	case string:
		return s.Table(v)
	default:
		return ToSess(s.db.Model(m))
	}
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
	err := defaultDB.Raw("SELECT EXISTS(?)", s.db).Debug().Scan(&exists).Error
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

func (s *Session) Update(value interface{}) (int64, error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return 0, fmt.Errorf("'value' must be struct, not '%T'", value)
	}
	r := s.db.Updates(value)
	return r.RowsAffected, r.Error
}

func (s *Session) UpdateColumn(column string, value interface{}) (int64, error) {
	r := s.db.Update(column, value)
	return r.RowsAffected, r.Error
}

func (s *Session) UpdateAttrs(attrs map[string]interface{}) (int64, error) {
	r := s.db.Updates(attrs)
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
	// ToSess 总是创建一个新 session
	// 如果直接返回 db，gorm2 的 statement 重用机制会继续保留己记录的 where 条件
	return &Session{db: db.Session(&gorm.Session{})}
}

func ToColName(name string) string {
	name = namingStrategy.ColumnName("", name)
	if i := strings.IndexByte(name, '.'); i >= 0 {
		name = name[i+1:]
	}
	return name
}

func Get() *Session {
	return ToSess(defaultDB)
}

func openDB(dsn string) error {
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

	mysqlDial := mysql.New(mysql.Config{
		DSN:               dsn,
		DefaultStringSize: 255,
	})
	db, err := gorm.Open(mysqlDial, &gorm.Config{
		NamingStrategy: namingStrategy,
		Logger: gormLogger.New(logs.Get(), gormLogger.Config{
			SlowThreshold:             slowThreshold,
			Colorful:                  false,
			IgnoreRecordNotFoundError: false,
			LogLevel:                  logLevel,
		}),
	})
	if err != nil {
		return err
	}

	if err = db.Callback().Create().Before("gorm:create").
		Register("my_before_create_hook", beforeCreateCallback); err != nil {
		return err
	}

	defaultDB = db
	//defaultDB = db.Session(&gorm.Session{
	//	Logger: gormLogger.New(logs.Get(), gormLogger.Config{
	//		SlowThreshold:             slowThreshold,
	//		Colorful:                  false,
	//		IgnoreRecordNotFoundError: false,
	//		LogLevel:                  logLevel,
	//	}),
	//})
	return nil
}

type CustomBeforeCreateInterface interface {
	CustomBeforeCreate(session *Session) error
}

func beforeCreateCallback(db *gorm.DB) {
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks && db.Statement.Schema.BeforeCreate {
		value := db.Statement.ReflectValue.Interface()
		if db.Statement.Schema.BeforeCreate {
			if i, ok := value.(CustomBeforeCreateInterface); ok {
				_ = db.AddError(i.CustomBeforeCreate(ToSess(db)))
			}
		}
	}
}

func Init(dsn string) {
	if err := openDB(dsn); err != nil {
		logs.Get().Fatalln(err)
	}
}
