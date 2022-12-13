package pgx

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun/schema"
)

type alterOption struct {
	add, drop, change bool      // column
	output            io.Writer // sql output
}

// AlterOption present add/drop/change table columns
type AlterOption func(opt *alterOption)

// WithAlterAdd add new column when table exist
func WithAlterAdd() AlterOption {
	return func(opt *alterOption) {
		opt.add = true
	}
}

// WithAlterDrop drop deprecated column when column exist
func WithAlterDrop() AlterOption {
	return func(opt *alterOption) {
		opt.drop = true
	}
}

// WithAlterChange alter column when column properties changed
func WithAlterChange() AlterOption {
	return func(opt *alterOption) {
		opt.change = true
	}
}

// WithAlterOutput set sql output file
func WithAlterOutput(w io.Writer) AlterOption {
	return func(opt *alterOption) {
		opt.output = w
	}
}

type PGYesOrNo bool

var (
	_ driver.Valuer = (*PGYesOrNo)(nil)
	_ sql.Scanner   = (*PGYesOrNo)(nil)
)

func (p *PGYesOrNo) Scan(src interface{}) (err error) {
	switch v := src.(type) {
	case string:
		switch v {
		case "YES":
			*p = true
		case "NO":
			*p = false
		default:
			return fmt.Errorf("unsupported data content: %s", v)
		}
	default:
		return fmt.Errorf("unsupported data type: %T", src)
	}
	return
}

func (p *PGYesOrNo) Value() (driver.Value, error) {
	if *p {
		return "YES", nil
	}
	return "NO", nil
}

func (p *PGYesOrNo) String() string {
	if p == nil || !*p {
		return "NO"
	}
	return "YES"
}

// tableColumn table columns in db
type tableColumn struct {
	schema.BaseModel `bun:"table:information_schema.columns"` // nolint

	ColumnName    string    `bun:"column_name"`
	IsNullable    PGYesOrNo `bun:"is_nullable"`
	DataType      string    `bun:"data_type"`
	ColumnDefault string    `bun:"column_default"`
}

type tableColumns []tableColumn

func getTableName(db IDB, model any) string {
	return db.NewSelect().Model(model).GetTableName()
}

func diffFieldWithColumn(fields []*schema.Field, cols tableColumns) (as []*schema.Field, ds tableColumns, err error) {
	// 1. drop
	for _, c := range cols {
		find := false
		for _, f := range fields {
			if c.ColumnName == f.Name {
				find = true
				break
			}
		}
		if !find {
			ds = append(ds, c)
		}
	}

	// 2. add
	for _, f := range fields {
		find := false
		for _, c := range cols {
			if c.ColumnName == f.Name {
				find = true
				break
			}
		}
		if !find {
			as = append(as, f)
		}
	}

	return
}

func AlterModel(ctx context.Context, db IDB, schema string, model any, opts ...AlterOption) (err error) {
	// init option
	var option alterOption
	for _, op := range opts {
		op(&option)
	}

	tbName := getTableName(db, model)
	exists, _ := db.NewSelect().Table("information_schema.tables").
		Where("table_schema=?", schema).
		Where("table_name=?", tbName).
		Exists(ctx)
	if !exists {
		logger().Infow("table not exists", "schema", schema, "table", tbName)
		return nil
	}

	// get table columns from db
	var cols tableColumns
	err = db.NewSelect().Model(&tableColumn{}).
		Where("table_schema=?", schema).
		Where("table_name=?", tbName).
		Scan(ctx, &cols)

	if err != nil {
		return
	}

	// get model fields from structure
	fields := db.Dialect().Tables().Get(reflect.TypeOf(model)).Fields

	// diff fields
	as, ds, err := diffFieldWithColumn(fields, cols)
	if err != nil {
		return
	}

	if option.add {
		if err = addColumnQuery(ctx, db, schema, tbName, as, option.output); err != nil {
			return
		}
	}
	if option.drop {
		if err = dropColumnQuery(ctx, db, schema, tbName, ds, option.output); err != nil {
			return
		}
	}
	// nolint
	if option.change {
		// TODO column change
	}

	return nil
}

func getColumnIndirectDef(indirectType reflect.Type) (colDef string) {
	switch indirectType {
	case reflect.TypeOf(time.Time{}):
		colDef = "DEFAULT CURRENT_TIMESTAMP"
	default:
		switch indirectType.Kind() {
		case reflect.Bool:
			colDef = "DEFAULT FALSE"
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
			reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			colDef = "DEFAULT 0"
		case reflect.Float32, reflect.Float64:
			colDef = "DEFAULT 0"
		case reflect.Array, reflect.Slice:
			colDef = "DEFAULT '[]'"
		case reflect.String:
			colDef = "DEFAULT ''"
		case reflect.Map, reflect.Struct:
			colDef = "DEFAULT '{}'"
		default:
			return ""
		}
	}
	return colDef
}

func columnDefaultWithName(sqlName string) string {
	switch strings.ToLower(sqlName) {
	case "timestamptz", "timestamp with time zone":
		return "DEFAULT CURRENT_TIMESTAMP"
	case "timestamp", "timestamp without time zone":
		return "DEFAULT LOCALTIMESTAMP"
	case "date":
		return "DEFAULT CURRENT_DATE"
	case "time with time zone":
		return "DEFAULT CURRENT_TIME"
	case "time", "time without time zone":
		return "DEFAULT LOCALTIME"
	case "boolean":
		return "DEFAULT FALSE"
	case "real", "double precision":
		return "DEFAULT 0"
	case "smallint", "integer", "bigint":
		return "DEFAULT 0"
	case "text", "varchar", "char", "character varying", "character":
		return "DEFAULT ''"
	case "json", "jsonb":
		return "DEFAULT '{}'"
	case "inet":
		return "DEFAULT '0.0.0.0'::inet"
	case "macaddr", "macaddr8":
		return "DEFAULT '00-00-00-00-00-00'::macaddr"
		// TODO: case array and array_type
	}
	return ""
}

func getColumnDefault(f *schema.Field) (colDef string, err error) {
	// use specify, type:name
	if f.UserSQLType != f.DiscoveredSQLType {
		colDef := columnDefaultWithName(f.UserSQLType)
		if colDef == "" {
			colDef = getColumnIndirectDef(f.IndirectType)
			if colDef == "" {
				return colDef, fmt.Errorf("field(%s) has no default value", f.GoName)
			}
		}
	} else {
		// use indirect type
		colDef = getColumnIndirectDef(f.IndirectType)
		if colDef == "" {
			return colDef, fmt.Errorf("field(%s) has no default value", f.GoName)
		}
	}
	return
}

func addColumnQuery(ctx context.Context, db IDB, schema, tbName string, as []*schema.Field, output io.Writer) (err error) {
	alter := "ALTER TABLE IF EXISTS %q.%q ADD IF NOT EXISTS %q %s %s %s;"

	for _, f := range as {
		pgTag := f.Tag
		// column type
		sqlType := strings.ToUpper(f.UserSQLType)
		// notnull
		sqlNotNull := ""
		if pgTag.HasOption("notnull") {
			sqlNotNull = "NOT NULL"
		}
		// default
		sqlDefault := ""
		if len(f.SQLDefault) > 0 {
			sqlDefault = "DEFAULT " + string(f.SQLDefault)
		}

		if len(sqlNotNull) > 0 && len(sqlDefault) == 0 {
			sqlDefault, err = getColumnDefault(f)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("get table(%s) field default error", tbName))
			}
		}

		err = execColumnQuery(ctx, db, output, alter,
			schema, tbName, f.Name, sqlType, sqlNotNull, sqlDefault)
		if err != nil {
			return
		}
	}
	return
}

// nolint
func dropColumnQuery(ctx context.Context, db IDB, schema, tbName string, ds tableColumns, output io.Writer) (err error) {
	alter := "ALTER TABLE IF EXISTS %q.%q DROP IF EXISTS %q;"

	for _, c := range ds {
		err = execColumnQuery(ctx, db, output, alter,
			schema, tbName, c.ColumnName)
		if err != nil {
			return
		}
	}
	return
}

func execColumnQuery(ctx context.Context, db IDB, output io.Writer, alter string, params ...interface{}) (err error) {
	alterQuery := fmt.Sprintf(alter, params...)
	if output != nil {
		_, err = output.Write(append([]byte(alterQuery), '\n'))
		if err != nil {
			logger().Infow("write fail", "err", err)
		}
		return
	}

	_, err = db.ExecContext(ctx, alterQuery)
	if err != nil {
		logger().Infow("alter table fail", "err", err)
	} else {
		logger().Infow("alter table done", "query", alterQuery)
	}

	return err
}
