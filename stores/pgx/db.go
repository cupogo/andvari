package pgx

import (
	"context"
	"database/sql"
	"io/fs"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/uptrace/bun/schema"

	"github.com/cupogo/andvari/models/oid"
)

type IConn = bun.IConn
type IDB = bun.IDB
type Tx = bun.Tx
type TxOptions = sql.TxOptions
type Ident = bun.Ident
type Safe = bun.Safe

type QueryBuilder = bun.QueryBuilder
type SelectQuery = bun.SelectQuery
type QueryAppender = schema.QueryAppender

var (
	ErrNoRows = sql.ErrNoRows
	In        = bun.In
)

const (
	crapSuffix = "_trash"

	defaultLimit = 20
)

var (
	lastSchema     string
	lastSchemaCrap string
)

func LastSchema() string {
	return lastSchema
}

func LastSchemaCrap() string {
	return lastSchemaCrap
}

type DB struct {
	*bun.DB

	scDft, scCrap string // default and trash schemas
	ftsConfig     string
	ftsEnabled    bool
}

func Open(dsn string, ftscfg string, debug bool) (*DB, error) {
	pgconn := pgdriver.NewConnector(pgdriver.WithDSN(dsn))
	pgcfg := pgconn.Config()

	sqldb := sql.OpenDB(pgconn)
	db := bun.NewDB(sqldb, pgdialect.New(), bun.WithDiscardUnknownColumns())

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		logger().Infow("connect fail", "addr", pgcfg.Addr, "db", pgcfg.Database, "user", pgcfg.User, "err", err)
		return nil, err
	}
	logger().Debugw("connected OK", "db", db.String(), "addr", pgcfg.Addr, "db", pgcfg.Database, "user", pgcfg.User)

	w := &DB{DB: db, scDft: pgcfg.User, scCrap: pgcfg.User + crapSuffix}
	lastSchema = w.scDft
	lastSchemaCrap = w.scCrap

	if debug {
		debugHook := &DebugHook{Verbose: true}
		db.AddQueryHook(debugHook)
	}

	w.ftsConfig = ftscfg
	w.ftsEnabled = CheckTsCfg(ctx, db, ftscfg)

	if err := EnsureSchema(ctx, db, w.scDft); err != nil {
		return nil, err
	}

	for _, name := range []string{"citext", "intarray", "btree_gin", "btree_gist", "pg_trgm"} {
		_ = EnsureExtension(ctx, db, name)
	}

	return w, nil
}

func (w *DB) CreateTables(ctx context.Context, dropIt bool, tables ...any) error {
	return CreateModels(ctx, w.DB, dropIt, tables...)
}

func (w *DB) Schema() string {
	return w.scDft
}

func (w *DB) SchemaCrap() string {
	return w.scCrap
}

func (w *DB) List(ctx context.Context, spec ListArg, dataptr any) (total int, err error) {
	q := w.NewSelect().Model(dataptr).Apply(spec.Sift)
	if spec.Deleted() {
		q.ModelTableExpr(w.scCrap + ".?TableName AS ?TableAlias")
	}
	return QueryPager(ctx, spec, q)
}

func (w *DB) OpDeleteOID(ctx context.Context, table string, id string) error {
	_, _id, err := oid.Parse(id)
	if err != nil {
		return err
	}
	return OpDeleteInTrans(ctx, w.DB, w.scDft, w.scCrap, table, _id)
}

func (w *DB) OpDeleteAny(ctx context.Context, table string, _id any) error {
	return OpDeleteInTrans(ctx, w.DB, w.scDft, w.scCrap, table, _id)
}

func (w *DB) OpUndeleteOID(ctx context.Context, table string, id string) error {
	_, _id, err := oid.Parse(id)
	if err != nil {
		return err
	}
	return OpUndeletedInTrans(ctx, w.DB, w.scDft, w.scCrap, table, _id)
}

func (w *DB) GetTsCfg() (string, bool) {
	return w.ftsConfig, w.ftsEnabled
}

func (w *DB) ApplyTsQuery(q *SelectQuery, kw, sty string, args ...string) (*SelectQuery, error) {
	return DoApplyTsQuery(w.ftsEnabled, w.ftsConfig, q, kw, sty, args...)
}

// nolint
func (w *DB) bulkExecAllFsSQLs(ctx context.Context) error {
	for _, dbfs := range alldbfs {
		if err := BulkFsSQLs(ctx, w.DB, dbfs); err != nil {
			return err
		}
	}
	return nil
}

func (w *DB) InitSchemas(ctx context.Context, dropIt bool) error {
	if err := w.CreateTables(ctx, dropIt, allmodels...); err != nil {
		return err
	}
	logger().Infow("inited schema", "tables", len(allmodels))

	return w.bulkExecAllFsSQLs(ctx)
}

func (w *DB) RunMigrations(ctx context.Context, mfs fs.FS) error {
	var migrations = migrate.NewMigrations()
	if err := migrations.Discover(mfs); err != nil {
		return err
	}
	migrator := migrate.NewMigrator(w.DB, migrations)
	if err := migrator.Init(ctx); err != nil {
		return err
	}
	group, err := migrator.Migrate(ctx)
	if err != nil {
		logger().Infow("migrate fail", "err", err)
		return nil
	}

	logger().Infow("migrated", "result", group.String())
	return nil
}
