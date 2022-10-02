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
	// pgOption, err := pg.ParseURL(dsn)
	// if err != nil {
	// 	logger().Warnw("parse db url failed", "err", err)
	// 	return nil, err
	// }
	pgcfg := pgconn.Config()

	w := &DB{}
	sqldb := sql.OpenDB(pgconn)
	db := bun.NewDB(sqldb, pgdialect.New(), bun.WithDiscardUnknownColumns())

	logger().Debugw("parsed", "addr", pgcfg.Addr, "db", pgcfg.Database, "user", pgcfg.User)
	w.scDft = pgcfg.User
	lastSchema = w.scDft
	// if w.scDft == "" {
	// 	logger().Fatalw("pg.user is empty in DSN")
	// 	return nil, err
	// }
	w.scCrap = w.scDft + crapSuffix
	lastSchemaCrap = w.scCrap

	// db := pg.Connect(pgOption)
	// if debug {
	// 	debugHook := &DebugHook{Verbose: true}
	// 	db.AddQueryHook(debugHook)
	// }
	ctx := context.Background()
	w.DB = db
	w.ftsConfig = ftscfg
	w.ftsEnabled = CheckTsCfg(ctx, db, ftscfg)

	_ = EnsureSchema(ctx, db, w.scDft)
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

// deprecated
func (w *DB) GetTsSpec() *TextSearchSpec {
	tcn, enl := w.GetTsCfg()
	tss := &TextSearchSpec{cfgname: tcn, enabled: enl}
	return tss
}

func (w *DB) ApplyTsQuery(q *SelectQuery, kw, sty string, args ...string) (*SelectQuery, error) {
	return DoApplyTsQuery(w.ftsEnabled, w.ftsConfig, q, kw, sty, args...)
}

func (w *DB) RunMigrations(ctx context.Context, mfs fs.FS, dir string) error {
	var migrations = migrate.NewMigrations()
	// migrations.
	// collection.SetTableName(w.scDft + ".gopg_migrations")
	if err := migrations.Discover(mfs); err != nil {
		return err
	}
	migrator := migrate.NewMigrator(w.DB, migrations)
	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}

	logger().Infow("migrated", "result", group.String())
	return nil
}
