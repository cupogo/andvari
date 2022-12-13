package pgx

import (
	"context"
	"fmt"

	"github.com/uptrace/bun/schema"
)

const (
	TableTypeBase = "BASE TABLE"
)

const (
	syncTrashSchemaSegment = "\n-- \n" +
		"-- Name: %s; Type: SCHEMA; Schema: - \n" +
		"-- \n"
	syncTrashTableSegment = "\n-- \n" +
		"-- Name: %s; Type: TABLE; Schema: %s \n" +
		"-- \n"
	syncTrashColumnSegment = "\n-- \n" +
		"-- Name: %s.%s; Type: COLUMN; Schema: %s \n" +
		"-- \n"
)

// table in db
type table struct {
	schema.BaseModel `bun:"table:information_schema.tables"` // nolint

	TableName        string    `bun:"table_name"`
	TableType        string    `bun:"table_type"`
	IsInsertableInfo PGYesOrNo `bun:"is_insertable_into"`
}

// nolint
func createTrashSchema(ctx context.Context, db IDB, schema string, option alterOption) (err error) {
	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q;", schema)

	if option.output != nil {
		comment := fmt.Sprintf(syncTrashSchemaSegment, schema)
		cq := append([]byte(comment), []byte(query)...)
		cq = append(cq, '\n')
		_, err = option.output.Write(cq)
	} else {
		_, err = db.ExecContext(ctx, query)
	}
	return
}

func syncTrashSchema(ctx context.Context, db IDB, defSchema, trashSchema string, opts ...AlterOption) (err error) {
	var option alterOption
	for _, op := range opts {
		op(&option)
	}

	exists, _ := db.NewSelect().Table("information_schema.schemata").
		Where("schema_name=?", defSchema).
		Exists(ctx)
	if !exists {
		logger().Infow("schema not exists", "schema", defSchema)
		return nil
	}

	exists, _ = db.NewSelect().Table("information_schema.schemata").
		Where("schema_name=?", trashSchema).
		Exists(ctx)
	if !exists {
		logger().Infow("schema not exists", "schema", trashSchema)
		return nil
	}

	// sync new tables from default schema to trash schema
	err = syncTrashTables(ctx, db, defSchema, trashSchema, option)
	if err != nil {
		return
	}

	return nil
}

func syncTrashTables(ctx context.Context, db IDB, defSchema, trashSchema string, option alterOption) (err error) {
	// get schema tables
	var (
		defTables, trashTables []table
		insertAble             = PGYesOrNo(true)
	)

	err = db.NewSelect().Model(&defTables).
		Where("table_schema=?", defSchema).
		Where("table_type=?", TableTypeBase).
		Where("is_insertable_into=?", insertAble.String()).
		Scan(ctx)
	if err != nil && err != ErrNoRows {
		logger().Infow("get tables", "schema", defSchema)
		return nil
	}

	err = db.NewSelect().Model(&trashTables).
		Where("table_schema=?", trashSchema).
		Where("table_type=?", TableTypeBase).
		Where("is_insertable_into=?", insertAble.String()).
		Scan(ctx)
	if err != nil && err != ErrNoRows {
		logger().Infow("get tables", "schema", trashSchema)
		return nil
	}

	for _, def := range defTables {
		find := false
		for _, trash := range trashTables {
			if def.TableName == trash.TableName {
				find = true
				break
			}
		}

		if find {
			// add/drop trash columns without new tables.
			err = syncTrashColumns(ctx, db, defSchema, trashSchema, def.TableName, option)
			if err != nil {
				return
			}
		}
	}

	return
}

// nolint
func addTrashTable(ctx context.Context, db IDB, defSchema, trashSchema, tbName string, option alterOption) (err error) {
	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %q.%q (LIKE %q.%q INCLUDING DEFAULTS, PRIMARY KEY (id));",
		trashSchema, tbName, defSchema, tbName)
	if option.output != nil {
		comment := fmt.Sprintf(syncTrashTableSegment, tbName, trashSchema)
		cq := append([]byte(comment), []byte(query)...)
		cq = append(cq, '\n')

		_, err = option.output.Write(cq)
	} else {
		_, err = db.ExecContext(ctx, query)
	}
	if err != nil {
		return
	}
	return nil
}

func syncTrashColumns(ctx context.Context, db IDB, defSchema, trashSchema, tbName string, option alterOption) (err error) {
	// get table columns from db
	var defCols tableColumns
	err = db.NewSelect().Model(&tableColumn{}).
		Where("table_schema=?", defSchema).
		Where("table_name=?", tbName).
		Scan(ctx, &defCols)

	if err != nil {
		return
	}

	// get table columns from db
	var trashCols tableColumns
	err = db.NewSelect().Model(&tableColumn{}).
		Where("table_schema=?", trashSchema).
		Where("table_name=?", tbName).
		Scan(ctx, &trashCols)

	if err != nil {
		return
	}

	// diff columns between trash and default
	as, ds := diffTrashColumns(defCols, trashCols)

	// add columns
	err = addTrashColumn(ctx, db, trashSchema, tbName, as, option)
	if err != nil {
		return
	}
	// drop columns
	err = dropTrashColumn(ctx, db, trashSchema, tbName, ds, option)
	if err != nil {
		return
	}

	return
}

func diffTrashColumns(defCols, trashCols tableColumns) (adds, drops tableColumns) {
	// add
	for i, def := range defCols {
		find := false
		for _, trash := range trashCols {
			if def.ColumnName == trash.ColumnName {
				find = true
				break
			}
		}

		if !find {
			adds = append(adds, defCols[i])
		}
	}

	// drop
	for i, trash := range trashCols {
		find := false
		for _, def := range defCols {
			if trash.ColumnName == def.ColumnName {
				find = true
				break
			}
		}

		if !find {
			drops = append(drops, trashCols[i])
		}
	}

	return
}

func addTrashColumn(ctx context.Context, db IDB,
	schema, tbName string, columns tableColumns, option alterOption) (err error) {
	for _, col := range columns {
		notNull := ""
		if !col.IsNullable {
			notNull = "NOT NULL"
		}

		def := ""
		if len(col.ColumnDefault) > 0 {
			def = "DEFAULT " + col.ColumnDefault
		}
		query := fmt.Sprintf("ALTER TABLE IF EXISTS %q.%q ADD IF NOT EXISTS %q %s %s %s;",
			schema, tbName, col.ColumnName, col.DataType, notNull, def)

		if option.output != nil {
			comment := fmt.Sprintf(syncTrashColumnSegment, tbName, col.ColumnName, schema)
			cq := append([]byte(comment), []byte(query)...)
			cq = append(cq, '\n')

			_, err = option.output.Write(cq)
		} else {
			_, err = db.ExecContext(ctx, query)
		}
		if err != nil {
			return
		}
	}

	return
}

func dropTrashColumn(ctx context.Context, db IDB,
	schema, tbName string, columns tableColumns, option alterOption) (err error) {
	for _, col := range columns {
		query := fmt.Sprintf("ALTER TABLE IF EXISTS %q.%q DROP IF EXISTS %q;",
			schema, tbName, col.ColumnName)

		if option.output != nil {
			comment := fmt.Sprintf(syncTrashColumnSegment, tbName, col.ColumnName, schema)
			cq := append([]byte(comment), []byte(query)...)
			cq = append(cq, '\n')

			_, err = option.output.Write(cq)
		} else {
			_, err = db.ExecContext(ctx, query)
		}
		if err != nil {
			return
		}
	}
	return
}
