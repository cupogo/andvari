package pgx

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/cupogo/andvari/models/field"
	"github.com/cupogo/andvari/models/oid"
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
	if _, err := db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS "+name+" WITH SCHEMA "+sc[0]); err != nil {
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
		logger().Errorw("create model failed", "name", query.GetTableName(), "err", err)
		return
	}
	logger().Debugw("create model", "name", query.GetTableName())
	return
}

func querySort(p Pager, q *SelectQuery) *SelectQuery {
	if rule := p.GetSort(); len(rule) > 1 {
		orders := strings.Split(rule, ",")
		// two fields at most
		if len(orders) > 2 {
			orders = orders[:2]
		}
		for _, order := range orders {
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
			if len(key) > 0 && p.CanSort(key) {
				if len(op) > 0 {
					q.OrderExpr(key + " " + op)
				} else {
					q.OrderExpr(key)
				}
			}
		}
	}
	return q
}

// QueryPager 根据分页参数进行查询
func QueryPager(ctx context.Context, p Pager, q *SelectQuery) (count int, err error) {
	q = querySort(p, q)
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
	if err != nil {
		logger().Infow("select failed", "pager", p, "err", err)
	} else {
		p.SetTotal(count)
	}
	if err != nil && err == sql.ErrNoRows {
		err = ErrNotFound
	}

	return
}

func ModelWithPK(ctx context.Context, db IDB, obj Model, columns ...string) (err error) {
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
		return ModelWithPK(ctx, db, obj, columns...)
	}

	logger().Infow("invalid id", "id", id, "name", db.NewSelect().Model(obj).GetTableName())
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

// DoInsert insert with ignore duplicate (optional)
func DoInsert(ctx context.Context, db IDB, obj Model, args ...any) error {
	// Call to saving hook
	if err := callToBeforeCreateHooks(obj); err != nil {
		return err
	}

	if dtf, ok := obj.(CreatedSetter); ok {
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
		unikey := field.ID
		if k, ok := args[0].(string); ok && obj.IsZeroID() {
			unikey = k
			args = args[1:]
		}
		q.On("CONFLICT (?) DO UPDATE", Ident(unikey))
		var foundUpd bool
		for _, arg := range args {
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
	if vs, ok := obj.(interface{ IsSerial() bool }); ok && vs.IsSerial() {
		q.Returning("id")
	} else {
		q.Returning("NULL")
	}

	if _, err := q.Exec(ctx); err != nil {
		logger().Infow("insert model fail", "name", q.GetTableName(), "obj", obj, "err", err)
		return err
	} else {
		logger().Debugw("insert model ok", "name", q.GetTableName(), "id", obj.GetID())
	}

	return callToAfterCreateHooks(obj)
}

func DoUpdate(ctx context.Context, db IDB, obj Model, columns ...string) error {

	// Call to saving hook
	if err := callToBeforeUpdateHooks(obj); err != nil {
		logger().Infow("before update model fail", "obj", obj, "err", err)
		return err
	}

	if vo, ok := obj.(Changeable); ok {
		if len(columns) > 0 {
			vo.SetChange(columns...)
		}
		if vo.CountChange() == 0 {
			logger().Infow("unchange", "id", obj.GetID())
			return nil
		}

		vo.SetChange(field.Updated)
		columns = vo.GetChanges()
	} else if len(columns) == 0 {
		logger().Infow("unchange", "id", obj.GetID())
		return nil
	}

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
		logger().Infow("update model fail", "name", q.GetTableName(),
			"obj", obj, "columns", columns, "err", err)
		return err
	} else {
		if ov, ok := obj.(Changeable); ok && ov.IsLog() && operateModelLogFn != nil {
			err = operateModelLogFn(ctx, GetModelName(q), field.ModelOperateTypeUpdate, obj)
			if err != nil {
				logger().Infow("update model operateModelLogFn", "name", q.GetTableName(), "err", err)
			}
		}
		logger().Debugw("update model ok", "name", q.GetTableName(),
			"id", obj.GetID(), "columns", columns)

	}
	if vo, ok := obj.(interface{ SetIsUpdate(v bool) }); ok {
		vo.SetIsUpdate(true)
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
		err = ModelWithPK(ctx, db, exist)
	} else if len(args) > 1 && utils.EnsureArgs(2, args[0], args[1]) {
		err = ModelWithUnique(ctx, db, exist, args[0], args[1])
	}

	if err == nil && !exist.IsZeroID() {
		csfn()
		if obj.CountChange() == 0 {
			return
		}
		err = DoUpdate(ctx, db, exist)
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

type MetaValueFunc func(ctx context.Context, id oid.OID) (any, error)

func OpModelMetaSet(ctx context.Context, mm ModelMeta, key string, id oid.OID, fn MetaValueFunc) error {
	if !id.IsZero() {
		if val, err := fn(ctx, id); err != nil {
			return err
		} else if !utils.IsZero(val) {
			logger().Debugw("set meta", key, val)
			mm.MetaSet(key, val)
			mm.SetChange(field.Meta)
		}
	}
	return nil
}

func FilterError(err error) error {
	if err == ErrNoRows {
		return ErrNotFound
	}
	if e, ok := err.(PGError); ok {
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
