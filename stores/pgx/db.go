package pgx

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io/fs"
	"os"
	"reflect"
	"runtime"
	"strconv"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/uptrace/bun/schema"
	"github.com/yalue/merged_fs"

	"github.com/cupogo/andvari/models/oid"
	"github.com/cupogo/andvari/stores/pgx/bunotel"
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
type PGError = pgdriver.Error

type QueryApplierFn = func(q *SelectQuery) *SelectQuery

var (
	ErrBadConn = driver.ErrBadConn
	ErrNoRows  = sql.ErrNoRows
	In         = bun.In
	Array      = pgdialect.Array
)

const (
	crapSuffix = "_trash"

	defaultLimit = 20
)

var (
	lastSchema     string
	lastSchemaCrap string
	lastFTScfg     string
	lastFTSok      bool
)

func LastSchema() string {
	return lastSchema
}

func LastSchemaCrap() string {
	return lastSchemaCrap
}

func LastFTSConfig() string {
	return lastFTScfg
}

func LastFTSEnabled() bool {
	return lastFTSok
}

type DB struct {
	*bun.DB

	scDft, scCrap string // default and trash schemas
	ftsCfg        string
	ftsOk         bool
}

func OpenDB(dsn string) (db *bun.DB, user string, err error) {
	pgconn := pgdriver.NewConnector(pgdriver.WithDSN(dsn))
	pgcfg := pgconn.Config()

	sqldb := sql.OpenDB(pgconn)
	patchPool(sqldb)

	db = bun.NewDB(sqldb, pgdialect.New(), bun.WithDiscardUnknownColumns())

	if err = db.Ping(); err != nil {
		logger().Infow("connect fail", "addr", pgcfg.Addr, "db", pgcfg.Database, "user", pgcfg.User, "err", err)
		return
	}
	logger().Debugw("connected OK", "db", db.String(), "addr", pgcfg.Addr, "db", pgcfg.Database, "user", pgcfg.User)

	patchHookOTEL(db, pgcfg.Database)
	patchHookDebug(db)

	user = pgcfg.User

	return
}

// Open database with ftscfg
func Open(dsn string, ftscfg string, _ ...bool) (*DB, error) {
	db, user, err := OpenDB(dsn)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	w := &DB{DB: db,
		scDft:  user,
		scCrap: user + crapSuffix,
		ftsCfg: ftscfg,
		ftsOk:  CheckTsCfg(ctx, db, ftscfg),
	}

	lastSchema = w.scDft
	lastSchemaCrap = w.scCrap

	if len(ftscfg) > 0 {
		lastFTScfg = w.ftsCfg
		lastFTSok = w.ftsOk
	}

	if err := EnsureSchema(ctx, db, w.scDft); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *DB) Schema() string {
	return w.scDft
}

func (w *DB) SchemaCrap() string {
	return w.scCrap
}

// Deprecated: use ListModel
func (w *DB) List(ctx context.Context, spec ListArg, dataptr any) (total int, err error) {
	return w.ListModel(ctx, spec, dataptr)
}
func (w *DB) ListModel(ctx context.Context, spec ListArg, dataptr any) (total int, err error) {
	q := QueryList(ctx, w, spec, dataptr)
	if spec.Deleted() {
		q.ModelTableExpr(w.scCrap + ".?TableName AS ?TableAlias")
	}
	if !spec.HasColumn() && !spec.HasExcludeColumn() {
		q = ApplyQueryContext(ctx, q)
	}

	return QueryPager(ctx, spec, q)
}

func (w *DB) GetModel(ctx context.Context, obj Model, id any, columns ...string) (err error) {
	if !obj.SetID(id) || obj.IsZeroID() {
		return ErrEmptyPK
	}
	q := w.NewSelect().Model(obj).WherePK()

	if len(columns) > 0 {
		q.Column(columns...)
	} else {
		q = ApplyQueryContext(ctx, q)
	}

	err = q.Scan(ctx)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return
}

func (w *DB) DeleteModel(ctx context.Context, obj ModelIdentity, id any) error {
	if !obj.SetID(id) || obj.IsZeroID() {
		return ErrEmptyPK
	}
	return w.DB.RunInTx(ctx, nil, func(ctx context.Context, tx Tx) error {
		return DoDeleteM(ctx, tx, w.scDft, w.scCrap, obj)
	})
}

func (w *DB) UndeleteModel(ctx context.Context, obj Model, id any) error {
	if !obj.SetID(id) || obj.IsZeroID() {
		return ErrEmptyPK
	}
	q := w.NewDelete().Model(obj)
	return OpUndeletedInTrans(ctx, w.DB, w.Schema(), w.SchemaCrap(), q.GetTableName(), obj.GetID())
}

// deprecated by DeleteModel
func (w *DB) OpDeleteOID(ctx context.Context, table string, id string) error {
	_, _id, err := oid.Parse(id)
	if err != nil {
		return err
	}
	return OpDeleteInTrans(ctx, w.DB, w.scDft, w.scCrap, table, _id)
}

// deprecated by DeleteModel
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
	return w.ftsCfg, w.ftsOk
}

// Deprecated: use TextSearchSpec.Sift
func (w *DB) ApplyTsQuery(q *SelectQuery, kw, sty string, args ...string) *SelectQuery {
	return DoApplyTsQuery(w.ftsOk, w.ftsCfg, q, kw, sty, args...)
}

// nolint
func (w *DB) bulkExecAllFsSQLs(ctx context.Context) (count int, err error) {
	for _, dbfs := range alldbfs {
		if n, err := BulkFsSQLs(ctx, w.DB, dbfs); err != nil {
			return 0, err
		} else {
			count += n
		}
	}
	return
}

func (w *DB) InitSchemas(ctx context.Context, dropIt bool) error {
	for _, name := range trustExt {
		_ = EnsureExtension(ctx, w.DB, name)
	}

	if err := CreateModels(ctx, w.DB, dropIt, allmodels...); err != nil {
		return err
	}
	count, err := w.bulkExecAllFsSQLs(ctx)
	logger().Infow("inited schema", "tables", len(allmodels), "sqls", count)
	return err
}

func (w *DB) SyncSchema(ctx context.Context, opts ...AlterOption) error {
	return syncTrashSchema(ctx, w.DB, w.Schema(), w.SchemaCrap(), opts...)
}

func (w *DB) AlterModels(ctx context.Context, opts ...AlterOption) error {
	schemas := []string{w.Schema(), w.SchemaCrap()}
	for i := 0; i < len(allmodels); i++ {
		for j := 0; j < len(schemas); j++ {
			if err := AlterModel(ctx, w.DB, schemas[j], allmodels[i], opts...); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *DB) RunMigrations(ctx context.Context, mfs ...fs.FS) error {
	if len(mfs) == 0 {
		mfs = alterfs
	}
	var migrations = migrate.NewMigrations()
	if err := migrations.Discover(merged_fs.MergeMultiple(mfs...)); err != nil {
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

type QueryBase interface {
	GetModel() bun.Model
	GetTableName() string
	Operation() string
}

func ModelNameByQ(q QueryBase) string {
	if md := q.GetModel(); md != nil {
		if tm, ok := md.(bun.TableModel); ok {
			return tm.Table().TypeName
		}
		if v, ok := md.(ModelIdentity); ok {
			return v.IdentityModel()
		}
	}

	return q.GetTableName()
}

func ModelName(m any) string {
	if v, ok := m.(ModelIdentity); ok {
		return v.IdentityModel()
	}
	typ := reflect.TypeOf(m).Elem()
	return indirectType(typ).Name()
}

func patchPool(sqldb *sql.DB) {
	if s, ok := os.LookupEnv("PGX_MAX_OPEN_X"); ok && len(s) > 0 {
		if x, err := strconv.Atoi(s); err == nil && x > 0 && x <= 4 {
			maxOpenConns := x * runtime.GOMAXPROCS(0)
			sqldb.SetMaxOpenConns(maxOpenConns)
			sqldb.SetMaxIdleConns(maxOpenConns)
			logger().Debugw("set max open = x * maxProcs", "x", x)
		}
	}
}

func patchHookOTEL(db *bun.DB, dbname string) {
	if s, ok := os.LookupEnv("PGX_BUN_OTEL"); ok && len(s) > 0 {
		if s == "1" || s == "2" {
			db.AddQueryHook(bunotel.NewQueryHook(
				bunotel.WithDBName(dbname),
				bunotel.WithFormattedQueries(s == "2"),
			))
		}
	}
}

func patchHookDebug(db *bun.DB) {
	if s, ok := os.LookupEnv("PGX_QUERY_DEBUG"); ok && len(s) > 0 {
		if x, err := strconv.ParseInt(s, 10, 32); err == nil && x > 0 {
			debugHook := &DebugHook{Verbose: x > 1}
			db.AddQueryHook(debugHook)
		}
	}
}

func indirectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
