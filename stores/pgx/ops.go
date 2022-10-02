package pgx

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/cupogo/andvari/models/field"
	"github.com/cupogo/andvari/utils"
)

func EnsureSchema(ctx context.Context, db IConn, name string) error {
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+name); err != nil {
		logger().Infow("create schema fail", "name", name, "err", err)
		return err
	}
	return nil
}

func EnsureExtension(ctx context.Context, db IConn, name string, sc ...string) error {
	if len(sc) == 0 || len(sc[0]) == 0 {
		sc = []string{"public"}
	}
	if _, err := db.ExecContext(context.TODO(), "CREATE EXTENSION IF NOT EXISTS "+name+" WITH SCHEMA "+sc[0]); err != nil {
		logger().Infow("create extension fail", "name", name, "err", err)
		return err
	}
	return nil
}

func CreateModels(ctx context.Context, db IDB, dropIt bool, tables ...any) error {
	for _, table := range tables {
		if err := CreateModel(ctx, db, table, dropIt); err != nil {
			return err
		}
	}

	return nil
}

func CreateModel(ctx context.Context, db IDB, model any, dropIt bool) (err error) {
	query := db.NewCreateTable().Model(model).IfNotExists()
	if dropIt {
		_, err = db.NewDropTable().Model(model).IfExists().Cascade().Exec(ctx)
		if err != nil {
			logger().Errorw("drop model failed", "model", model, "err", err)
			return
		}
	}

	_, err = query.Exec(ctx)
	if err != nil {
		logger().Errorw("create model failed", "model", model, "err", err)
		return
	}
	logger().Debugw("create model", "name", query.GetTableName())
	return
}

// QueryPager 根据分页参数进行查询
func QueryPager(ctx context.Context, p Pager, q *SelectQuery) (count int, err error) {
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
		count, err = q.Limit(limit).Offset(skip).ScanAndCount(ctx)
	} else if limit < 0 {
		count, err = q.Count(ctx)
	} else {
		err = q.Scan(ctx)
	}
	if err != nil && err == sql.ErrNoRows {
		err = ErrNotFound
	} else {
		p.SetTotal(count)
	}
	if err != nil {
		logger().Infow("select failed", "pager", p, "err", err)
	}

	return
}

func ModelWherePK(ctx context.Context, db IDB, obj Model, columns ...string) (err error) {
	if obj.IsZeroID() {
		return ErrEmptyPK
	}

	err = db.NewSelect().Model(obj).Column(columns...).WherePK().Scan(ctx)

	if err == sql.ErrNoRows {
		logger().Debugw("get model where pk no rows", "objID", obj.GetID())
		return ErrNotFound
	}
	if err != nil {
		logger().Warnw("get model where pk failed", "objID", obj.GetID(), "err", err)
		return
	}
	return
}

func ModelWithPKID(ctx context.Context, db IDB, obj Model, id any, columns ...string) error {
	if obj.SetID(id) {
		return ModelWherePK(ctx, db, obj, columns...)
	}

	logger().Infow("invalid id", "id", id)
	return fmt.Errorf("invalid id: '%+v'", id)
}

func ModelWithUnique(ctx context.Context, db IDB, obj Model, key string, val any) error {
	if val == nil || val == 0 || val == "" {
		logger().Infow("empty param", "key", key, "val", val)
		return ErrEmptyKey
	}
	err := db.NewSelect().Model(obj).Where("? = ?", Ident(key), val).Limit(1).Scan(ctx)
	if err == sql.ErrNoRows {
		logger().Debugw("get model with key no rows", "key", key, "val", val)
		return ErrNotFound
	}
	if err != nil {
		logger().Warnw("get model with key failed", "key", key, "val", val, "err", err)
		return err
	}
	return nil
}

// DoInsert insert with ignore duplicate (force)
func DoInsert(ctx context.Context, db IDB, obj Model, args ...any) error {
	// Call to saving hook
	if err := callToBeforeCreateHooks(obj); err != nil {
		return err
	}

	if dtf, ok := obj.(interface{ SetCreated(ts any) bool }); ok {
		if ts, ok := CreatedFromContext(ctx); ok && ts > 0 {
			if dtf.SetCreated(ts) {
				logger().Infow("seted createAt ok", "ts", ts)
			} else {
				logger().Infow("seted createAt fail", "ts", ts)
			}
		}
	}

	q := db.NewInsert().Model(obj)
	argc := len(args)
	if argc > 0 {
		q.On("CONFLICT (?) DO UPDATE", Ident(field.ID))
		var foundUpd bool
		for i, arg := range args {
			if b, ok := arg.(bool); ok && b && i == 0 {
				q.Set("?0 = EXCLUDED.?0", Ident(field.Updated))
				foundUpd = true
				break
			}
			if a, ok := arg.(string); ok {
				q.Set("?0 = EXCLUDED.?0", Ident(a))
				if a == field.Updated {
					foundUpd = true
				}
			}
		}
		if !foundUpd {
			q.Set("?0 = EXCLUDED.?0", Ident(field.Updated))
		}

	}

	if _, err := q.Exec(ctx); err != nil {
		logger().Infow("insert model fail", "obj", obj, "err", err)
		return err
	} else {
		logger().Debugw("insert model ok", "id", obj.GetID(), "name", q.GetTableName())
	}

	return callToAfterCreateHooks(obj)
}

func DoUpdate(ctx context.Context, db IDB, obj ModelChangeable, columns ...string) error {
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

	q := db.NewUpdate().Model(obj).Column(columns...)
	if tso, ok := obj.(TextSearchable); ok {
		cfg := tso.GetTsConfig()
		cols := tso.GetTsColumns()
		if len(cfg) > 0 && len(cols) > 0 {
			for _, co := range columns {
				q.Set(co + " = ?" + co)
			}
			q.Set("ts_vec = to_tsvector(?, jsonb_build_array("+strings.Join(cols, ",")+"))", cfg)
		}
	}
	if _, err := q.WherePK().Exec(ctx); err != nil {
		logger().Infow("update model fail", "obj", obj, "columns", columns, "err", err)
		return err
	} else {
		logger().Debugw("update model ok", "id", obj.GetID(),
			"name", q.GetTableName(),
			"columns", columns)
	}

	return callToAfterUpdateHooks(obj)
}

func StoreSimple(ctx context.Context, db IDB, obj ModelChangeable, columns ...string) error {
	if obj.IsZeroID() {
		return DoInsert(ctx, db, obj)
	}

	return DoUpdate(ctx, db, obj, columns...)
}

type columnsFn func() []string

func StoreWithCall(ctx context.Context, db IDB, exist, obj ModelChangeable, csfn columnsFn, args ...string) (isn bool, err error) {
	if !obj.IsZeroID() {
		exist.SetID(obj.GetID())
		err = ModelWherePK(ctx, db, exist)
	} else if len(args) > 1 && utils.EnsureArgs(2, args[0], args[1]) {
		err = ModelWithUnique(ctx, db, exist, args[0], args[1])
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

func DoDelete(ctx context.Context, db IDB, table string, _id any) error {
	return DoDeleteT(ctx, db, LastSchema(), LastSchemaCrap(), table, _id)
}

func OpDeleteInTrans(ctx context.Context, db IDB, scDft, scCrap string, table string, _id any) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx Tx) error {
		return DoDeleteT(ctx, tx, scDft, scCrap, table, _id)
	})
}

// DoDeleteT call sp to do affect delete with table and id // TODO: id as int64
func DoDeleteT(ctx context.Context, db IDB, scDft, scCrap string, table string, _id any) error {
	var ret int
	err := db.NewRaw("SELECT op_affect_delete(?, ?, ?, ?)", scDft, scCrap, table, _id).Scan(ctx, &ret)
	if err != nil {
		logger().Infow("delete fail", "table", table, "id", _id, "err", err)
	} else {
		logger().Infow("delete ok", "table", table, "id", _id, "ret", ret)
	}
	return err
}

func OpUndeletedInTrans(ctx context.Context, db IDB, scDft, scCrap string, table string, _id any) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx Tx) error {
		return DoUndeleteT(ctx, tx, scDft, scCrap, table, _id)
	})
}

// DoUndeleteT call sp to do affect undelete with table and id
func DoUndeleteT(ctx context.Context, db IDB, scDft, scCrap string, table string, _id any) error {
	var ret int
	err := db.NewRaw("SELECT op_affect_undelete(?, ?, ?, ?)", scDft, scCrap, table, _id).Scan(ctx, &ret)
	if err != nil {
		logger().Infow("undelete fail", "table", table, "id", _id, "err", err)
	} else {
		logger().Infow("undelete ok", "table", table, "id", _id, "ret", ret)
	}
	return err
}

// func FilterError(err error) error {
// 	if e, ok := err.(pg.Error); ok {
// 		switch e.Field('C') {
// 		case "23502":
// 			return ErrEmptyKey
// 		case "23505":
// 			return ErrDuplicate
// 		}
// 		return ErrInternal
// 	}
// 	return err
// }
