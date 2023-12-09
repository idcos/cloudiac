package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
)

type SqlAdapterPlugin struct {
}

func (s *SqlAdapterPlugin) Name() string {
	return "sqlAdapterPlugin"
}
func (s *SqlAdapterPlugin) Initialize(db *gorm.DB) error {
	if err := db.Callback().Query().Replace("gorm:query", Query); err != nil {
		return err
	}
	if err := db.Callback().Row().Replace("gorm:row", RowQuery); err != nil {
		return err
	}
	return nil
}

func BuildQuerySQL(db *gorm.DB) {
	callbacks.BuildQuerySQL(db)
	sql := db.Statement.SQL.String()
	sql = GetDriver().SQLEnhance(sql)
	db.Statement.SQL.Reset()
	db.Statement.SQL.WriteString(sql)
}

func Query(db *gorm.DB) {
	if db.Error == nil {
		BuildQuerySQL(db)
		if !db.DryRun && db.Error == nil {
			rows, err := db.Statement.ConnPool.QueryContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
			if err != nil {
				db.AddError(err)
				return
			}
			defer func() {
				db.AddError(rows.Close())
			}()
			gorm.Scan(rows, db, 0)
		}
	}
}

func RowQuery(db *gorm.DB) {
	if db.Error == nil {
		BuildQuerySQL(db)
		if db.DryRun {
			return
		}

		if isRows, ok := db.Get("rows"); ok && isRows.(bool) {
			db.Statement.Settings.Delete("rows")
			db.Statement.Dest, db.Error = db.Statement.ConnPool.QueryContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
		} else {
			db.Statement.Dest = db.Statement.ConnPool.QueryRowContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
		}

		db.RowsAffected = -1
	}
}
