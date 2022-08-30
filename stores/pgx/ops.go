package pgx

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"

	"hyyl.xyz/cupola/andvari/models/field"
	"hyyl.xyz/cupola/andvari/utils"
)

type ormDB = orm.DB
type ormQuery = orm.Query
type ormResult = orm.Result
type pgDB = pg.DB
type pgTx = pg.Tx
type pgIdent = pg.Ident

func EnsureSchema(db *pgDB, scDft string) error {
	if _, err := db.Exec("CREATE SCHEMA IF NOT EXISTS " + scDft); err != nil {
		logger().Infow("create schema fail", "sc", scDft, "err", err)
		return err
	}
	return nil
}

func CreateModels(db *pgDB, dropIt bool, tables ...any) error {
	for _, table := range tables {
		if err := CreateModel(db, table, dropIt); err != nil {
			return err
		}
	}

	return nil
}

// QueryPager 根据分页参数进行查询
func QueryPager(p Pager, q *ormQuery) (count int, err error) {
	if order := p.GetSort(); len(order) > 1 {
		var key, op string
		if b, a, ok := strings.Cut(order, " "); ok {
			op = strings.ToUpper(a)
			if op == "DESC" || op == "ASC" {
				key = b
			}
		} else if strings.Index(order, "-") == 0 { // -createdAt
			key = order[1:]
			op = "DESC"
		} else {
			key = order
		}
		if p.CanSort(key) {
			if len(op) > 0 {
				q.Order(key + " " + op)
			} else {
				q.Order(key)
			}

		}
	}
	limit := p.GetLimit()
	if p.GetPage() > 0 && limit == 0 {
		limit = defaultLimit
	}

	if limit > 0 {
		skip := p.GetSkip()
		if skip == 0 && p.GetPage() > 0 {
			skip = (p.GetPage() - 1) * limit
		}
		count, err = q.Limit(limit).Offset(skip).SelectAndCount()
	} else if limit < 0 {
		count, err = q.Count()
	} else {
		err = q.Select()
	}
	if err != nil && err == pg.ErrNoRows {
		err = ErrNotFound
	} else {
		p.SetTotal(count)
	}
	if err != nil {
		logger().Infow("select failed", "pager", p, "err", err)
	}

	return
}

func ModelWherePK(ctx context.Context, db orm.DB, obj Model, columns ...string) (err error) {
	if obj.IsZeroID() {
		return ErrEmptyPK
	}

	err = db.ModelContext(ctx, obj).Column(columns...).WherePK().Select()

	if err == pg.ErrNoRows {
		logger().Debugw("get model where pk no rows", "objID", obj.GetID())
		return ErrNotFound
	}
	if err != nil {
		logger().Warnw("get model where pk failed", "objID", obj.GetID(), "err", err)
		return
	}
	return
}

func ModelWithPKID(ctx context.Context, db ormDB, obj Model, id any, columns ...string) error {
	if obj.SetID(id) {
		return ModelWherePK(ctx, db, obj, columns...)
	}

	logger().Infow("invalid id", "id", id)
	return fmt.Errorf("invalid id: '%+v'", id)
}

func ModelWithUnique(db ormDB, obj Model, key string, val any) error {
	if val == nil || val == 0 || val == "" {
		logger().Infow("empty param", "key", key, "val", val)
		return ErrEmptyKey
	}
	err := db.Model(obj).Where("? = ?", pg.Ident(key), val).Limit(1).Select()
	if err == pg.ErrNoRows {
		logger().Debugw("get model with key no rows", "key", key, "val", val)
		return ErrNotFound
	}
	if err != nil {
		logger().Warnw("get model with key failed", "key", key, "val", val, "err", err)
		return err
	}
	return nil
}

func CheckTsCfg(db ormDB, ftsConfig string) bool {
	var ret int
	_, err := db.QueryOne(pg.Scan(&ret), "SELECT oid FROM pg_ts_config WHERE cfgname = ?", ftsConfig)
	if err == nil {
		if ret > 0 {
			logger().Infow("fts checked ok", "ts cfg", ftsConfig)
			return true
		}
	} else {
		logger().Infow("select ts config fail", "tscfg", ftsConfig, "err", err)
	}
	return false
}

func CreateModel(db *pg.DB, model any, dropIt bool) (err error) {
	var opt orm.CreateTableOptions
	query := db.Model(model)
	if dropIt {
		err = query.DropTable(&orm.DropTableOptions{
			IfExists: true,
			Cascade:  true,
		})

		if err != nil {
			logger().Errorw("drop model failed", "model", model, "err", err)
			return
		}
	} else {
		opt.IfNotExists = true
	}

	err = query.CreateTable(&opt)
	if err != nil {
		logger().Errorw("create model failed", "model", model, "err", err)
		return
	}
	logger().Infow("create model", "name", query.TableModel().Table().SQLName)
	return
}

func StoreSimple(ctx context.Context, db ormDB, obj Model, columns ...string) error {
	if obj.IsZeroID() {
		return DoInsert(ctx, db, obj)
	}

	return DoUpdate(ctx, db, obj, columns...)
}

type columnsFn func() []string

func StoreWithCall(ctx context.Context, db ormDB, exist, obj Model, csfn columnsFn, args ...string) (isn bool, err error) {
	if !obj.IsZeroID() {
		exist.SetID(obj.GetID())
		err = ModelWherePK(ctx, db, exist)
	} else if len(args) > 1 && utils.EnsureArgs(2, args[0], args[1]) {
		err = ModelWithUnique(db, exist, args[0], args[1])
	}

	if err == nil && !exist.IsZeroID() {
		cs := csfn()
		if len(cs) == 0 { // unchanged
			return
		}
		err = DoUpdate(ctx, db, exist, cs...)
	} else {
		isn = true
		err = DoInsert(ctx, db, obj)
	}
	return
}

// DoInsert insert with ignore duplicate (force)
func DoInsert(ctx context.Context, db ormDB, obj Model, args ...any) error {
	// Call to saving hook
	if err := callToBeforeCreateHooks(obj); err != nil {
		return err
	}

	q := db.ModelContext(ctx, obj)
	if len(args) > 0 {
		if b, ok := args[0].(bool); ok && b {
			if v, ok := obj.(updatable); ok {
				q.OnConflict("(?) DO UPDATE", pgIdent(field.ID)).Set(
					"? = ?", pgIdent(field.Updated), v.GetUpdated(),
				)
			}
		} else if utils.EnsureArgs(2, args...) {
			if _f, ok := args[0].(string); ok {
				q.OnConflict("(?) DO UPDATE", pgIdent(field.ID)).Set(
					"? = ?", pgIdent(_f), args[1])
			}

		}

	}

	if _, err := q.Insert(); err != nil {
		logger().Infow("insert model fail", "obj", obj, "err", err)
		return err
	} else {
		logger().Debugw("insert model ok", "id", obj.GetID(), "name", q.TableModel().Table().SQLName)
	}

	return callToAfterCreateHooks(obj)
}

func DoUpdate(ctx context.Context, db ormDB, obj Model, columns ...string) error {
	if len(columns) > 0 {
		obj.SetChange(columns...)
	}
	// Call to saving hook
	if err := callToBeforeUpdateHooks(obj); err != nil {
		logger().Infow("before update model fail", "obj", obj, "err", err)
		return err
	}

	if obj.CountChange() == 0 {
		logger().Infow("unchange", "id", obj.GetID())
		return nil
	}
	obj.SetChange(field.Updated)
	columns = obj.GetChanges()

	if _, err := db.ModelContext(ctx, obj).Column(columns...).WherePK().Update(); err != nil {
		logger().Infow("update model fail", "obj", obj, "columns", columns, "err", err)
		return err
	}

	return callToAfterUpdateHooks(obj)
}

func DoDelete(ctx context.Context, db ormDB, table string, _id any) error {
	return DoDeleteT(ctx, db, LastSchema(), LastSchemaCrap(), table, _id)
}

func OpDeleteInTrans(ctx context.Context, db *pgDB, scDft, scCrap string, table string, _id any) error {
	return db.RunInTransaction(ctx, func(tx *pgTx) error {
		return DoDeleteT(ctx, db, scDft, scCrap, table, _id)
	})
}

// DoDeleteT call sp to do affect delete with table and id // TODO: id as int64
func DoDeleteT(ctx context.Context, db ormDB, scDft, scCrap string, table string, _id any) error {
	var ret int
	_, err := db.QueryOneContext(ctx, pg.Scan(&ret), "SELECT op_affect_delete(?, ?, ?, ?)", scDft, scCrap, table, _id)
	if err != nil {
		logger().Infow("delete fail", "table", table, "err", err)
	} else {
		logger().Infow("delete ok", "table", table, "ret", ret)
	}
	return err
}

func OpUndeletedInTrans(ctx context.Context, db *pgDB, scDft, scCrap string, table string, _id any) error {
	return db.RunInTransaction(ctx, func(tx *pgTx) error {
		return DoUndeleteT(ctx, db, scDft, scCrap, table, _id)
	})
}

// DoUndeleteT call sp to do affect undelete with table and id
func DoUndeleteT(ctx context.Context, db ormDB, scDft, scCrap string, table string, _id any) error {
	var ret int
	_, err := db.QueryOneContext(ctx, pg.Scan(&ret), "SELECT op_affect_undelete(?, ?, ?, ?)", scDft, scCrap, table, _id)
	if err != nil {
		logger().Infow("undelete fail", "table", table, "err", err)
	} else {
		logger().Infow("undelete ok", "table", table, "ret", ret)
	}
	return err
}

func FilterError(err error) error {
	if e, ok := err.(pg.Error); ok {
		switch e.Field('C') {
		case "23502":
			return ErrEmptyKey
		case "23505":
			return ErrDuplicate
		}
		return ErrInternal
	}
	return err
}
