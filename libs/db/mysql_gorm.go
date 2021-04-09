package db

import (
	"database/sql"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"cloudiac/configs"
	"cloudiac/consts/e"
	"cloudiac/utils/logs"
)

var db *gorm.DB
var logger logs.Logger

type Session struct {
	db *gorm.DB
}

func (s *Session) Begin() *Session {
	return ToSess(s.db.Begin())
}

func (s *Session) DB() *gorm.DB {
	return s.db
}

func (s *Session) New() *Session {
	return ToSess(s.db.New())
}

func (s *Session) AddIndex(indexName string, columns ...string) error {
	return s.DB().AddIndex(indexName, columns...).Error
}

func (s *Session) AddUniqueIndex(indexName string, columns ...string) error {
	return s.DB().AddUniqueIndex(indexName, columns...).Error
}

func (s *Session) RemoveIndex(table string, indexName string) error {
	if !s.DB().Dialect().HasIndex(table, indexName) {
		return nil
	}

	return s.DB().Table(table).RemoveIndex(indexName).Error
}

func (s *Session) DropColumn(table string, columns ...string) error {
	dbTable := s.DB().Table(table)
	logger := logger.WithField("func", "DropColumn")
	for _, col := range columns {
		err := dbTable.DropColumn(col).Error
		if err != nil {
			if e.IsMysqlErr(err, e.MysqlDropColOrKeyNotExists) {
				logger.Infof("column '%s' not exists, ignore", col)
				continue
			}
			return err
		}
	}
	return nil
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

func (s *Session) Expr() *gorm.SqlExpr {
	return s.db.QueryExpr()
}

// Expr generate raw SQL expression, for example:
//     DB.Model(&product).Update("price", gorm.Expr("price * ? + ?", 2, 100))
func (s *Session) GormExpr(expression string, args ...interface{}) *gorm.SqlExpr {
	return gorm.Expr(expression, args...)
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

func (s *Session) Order(value interface{}, reorder ...bool) *Session {
	return ToSess(s.db.Order(value, reorder...))
}

func (s *Session) Group(query string) *Session {
	return ToSess(s.db.Group(query))
}

func (s *Session) Limit(n interface{}) *Session {
	return ToSess(s.db.Limit(n))
}

func (s *Session) Offset(n interface{}) *Session {
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
	err := s.db.New().Raw("SELECT EXISTS(?)", s.Expr()).Row().Scan(&exists)
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

func (s *Session) Update(attrs ...interface{}) (int64, error) {
	r := s.db.Update(attrs...)
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
将指定的 field 与 q 比较大小

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

func openDB(args string) error {
	var err error
	db, err = gorm.Open("mysql", args)
	return err
}

func ToSess(db *gorm.DB) *Session {
	return &Session{db: db}
}

func ToColName(name string) string {
	return gorm.ToColumnName(name)
}

type sqlLogger struct {
	logger logs.Logger
}

func (sqlLogger) Print(v ...interface{}) {
	var (
		typ      string
		fileLine string
		ok       bool
	)

	if len(v) < 3 {
		goto end
	}

	typ, ok = v[0].(string)
	if !ok {
		goto end
	}

	fileLine, ok = v[1].(string)
	if !ok {
		goto end
	}

	// see gorm/main.go:DB.logs()
	if typ == "sql" || typ == "logs" {
		for i := 7; i < 32; i++ {
			_, file, line, ok := runtime.Caller(i)
			if ok && fmt.Sprintf("%s:%d", file, line) == fileLine {
				_, file, line, ok = runtime.Caller(i + 1)
				if ok {
					v[1] = fmt.Sprintf("%s:%d", file, line)
				}
				break
			}
		}
	}

end:
	logger.Debugln(gorm.LogFormatter(v...)...)
}

func Get() *Session {
	if db == nil {
		conf := configs.Get()
		if err := openDB(conf.Mysql); err != nil {
			logger.Panicln(err)
		}
		db.SetLogger(sqlLogger{logger})
	}

	return ToSess(db)
}

func Init() {
	logger = logs.Get()
	Get()
}
