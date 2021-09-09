// Copyright 2021 CloudJ Company Limited. All rights reserved.

package db

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/soft_delete"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"cloudiac/portal/consts/e"
	"cloudiac/utils/logs"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var (
	defaultDB      *gorm.DB
	namingStrategy = schema.NamingStrategy{}
)

type SoftDeletedAt uint

func (v SoftDeletedAt) QueryClauses(f *schema.Field) []clause.Interface {
	return soft_delete.DeletedAt(v).QueryClauses(f)
}

func (v SoftDeletedAt) DeleteClauses(f *schema.Field) []clause.Interface {
	return soft_delete.DeletedAt(v).DeleteClauses(f)
}

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

	err := s.db.Exec(fmt.Sprintf("CREATE UNIQUE INDEX `%s` ON `%s` (%s)",
		indexName, stmt.Table, strings.Join(columns, ","))).Error
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

func (s *Session) DropColumn(model interface{}, columns ...string) error {
	migrator := s.db.Migrator()
	for _, col := range columns {
		if migrator.HasColumn(model, col) {
			if err := migrator.DropColumn(model, col); err != nil {
				return err
			}
		}
	}
	return nil
}

// ModifyModelColumn 基于 gorm:"" tag 同步字段类型定义
func (s *Session) ModifyModelColumn(model interface{}, column string) error {
	if !s.isModel(model) {
		return fmt.Errorf("'model' must be a 'struct', not '%T'", model)
	}
	return s.db.Migrator().AlterColumn(model, column)
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
	qs := s.autoLazySelect()
	return qs.db
}

func (s *Session) Raw(sql string, values ...interface{}) *Session {
	// FIXME: gorm driver bugs
	// gorm@v1.21.12~14: statement.go +204
	//   subdb.Statement.Vars = stmt.Vars
	// when values is a *DB (usually a sub query), the statement vars will be appended twice
	// workaround:
	//  check if sql variables wanted matched the vars count x 2, then remove the duplicated vars
	ss := ToSess(s.db.Raw(sql, values...))
	// remove duplicated vars
	if len(ss.db.Statement.Vars) > 0 && strings.Count(ss.db.Statement.SQL.String(), "?")*2 == len(ss.db.Statement.Vars) {
		logs.Get().Warnf("gorm bugs: duplicate vars, sql: %s", ss.db.Statement.SQL.String())
		ss.db.Statement.Vars = ss.db.Statement.Vars[:len(ss.db.Statement.Vars)/2]
	}

	return ss
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

func (s *Session) Having(query interface{}, args ...interface{}) *Session {
	return ToSess(s.db.Having(query, args...))
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
	qs := s.autoLazySelect()
	err = qs.db.Count(&cnt).Error
	return
}

func (s *Session) Exists() (bool, error) {
	exists := false
	err := s.New().Raw("SELECT EXISTS(?)", s.db).Scan(&exists)
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

func (s *Session) isModel(m interface{}) bool {
	rv := reflect.ValueOf(m)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return rv.Kind() == reflect.Struct
}

func (s *Session) Update(value interface{}) (int64, error) {
	if !s.isModel(value) {
		return 0, fmt.Errorf("'value' must be a 'struct', not '%T'", value)
	}
	r := s.db.Updates(value)
	return r.RowsAffected, r.Error
}

func (s *Session) UpdateColumn(column string, value interface{}) (int64, error) {
	r := s.db.Update(ToColName(column), value)
	return r.RowsAffected, r.Error
}

func (s *Session) UpdateAttrs(attrs map[string]interface{}) (int64, error) {
	newAttrs := make(map[string]interface{}, len(attrs))
	for k, v := range attrs {
		newAttrs[ToColName(k)] = v
	}
	r := s.db.Updates(newAttrs)
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

	if err = db.Callback().Create().Before("gorm:before_create").
		Register("my_before_create_hook", beforeCreateCallback); err != nil {
		return err
	}

	defaultDB = db
	return nil
}

type CustomBeforeCreateInterface interface {
	CustomBeforeCreate(session *Session) error
}

// callMethod gorm.io/gorm@v1.21.12/callbacks/callmethod.go
func callMethod(db *gorm.DB, fc func(value interface{}, tx *gorm.DB) bool) {
	tx := db.Session(&gorm.Session{NewDB: true})
	if called := fc(db.Statement.ReflectValue.Interface(), tx); !called {
		switch db.Statement.ReflectValue.Kind() {
		case reflect.Slice, reflect.Array:
			db.Statement.CurDestIndex = 0
			for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
				fc(reflect.Indirect(db.Statement.ReflectValue.Index(i)).Addr().Interface(), tx)
				db.Statement.CurDestIndex++
			}
		case reflect.Struct:
			fc(db.Statement.ReflectValue.Addr().Interface(), tx)
		}
	}
}

func beforeCreateCallback(db *gorm.DB) {
	// 这里不需要判断 db.Statement.Schema.BeforeCreate,
	// Schema.BeforeCreate 只是用于判断 model 是否定义了 BeforeCreate(*gorm.DB) 方法，如果未定义则该值为 false。
	// gorm 定义 Schema.BeforeCreate 是为了避免每次执行都进行接口断言？
	// Schema.BeforeCreate 赋值: gorm.io/gorm@v1.21.12/schema/schema.go:229
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks {
		callMethod(db, func(value interface{}, db *gorm.DB) (called bool) {
			if i, ok := value.(CustomBeforeCreateInterface); ok {
				called = true
				_ = db.AddError(i.CustomBeforeCreate(ToSess(db)))
			}
			return called
		})
	}
}

func Init(dsn string) {
	if err := openDB(dsn); err != nil {
		logs.Get().Fatalln(err)
	}
}
