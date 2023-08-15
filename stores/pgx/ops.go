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
		logger().Errorw("create model failed", "name", GetTableName(query, model), "err", err)
		return
	}
	logger().Debugw("create model", "name", GetTableName(query, model))
	return
}

func ApplyQuerySort(p Sortable, q *SelectQuery) *SelectQuery {
	if rule := p.GetSort(); len(rule) > 1 {
		for _, order := range strings.Split(rule, ",") {
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
	q = ApplyQuerySort(p, q)
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

	q := db.NewSelect().Model(obj).Column(columns...).WherePK()
	err = q.Scan(ctx)
	if err != nil {
		if err == ErrNoRows {
			logger().Debugw("get model where pk no rows", "name", GetTableName(q, obj), "objID", obj.GetID())
			return ErrNotFound
		}
		logger().Warnw("get model where pk failed", "objID", obj.GetID(), "err", err)
		if err == ErrBadConn {
			panic(err)
		}
		return
	}
	return
}

func ModelWithPKID(ctx context.Context, db IDB, obj Model, id any, columns ...string) error {
	if obj.SetID(id) {
		return ModelWithPK(ctx, db, obj, columns...)
	}

	logger().Infow("invalid id", "id", id, "name", GetTableName(db.NewSelect().Model(obj), obj))
	return fmt.Errorf("invalid id: '%+v'", id)
}

func ModelWithUnique(ctx context.Context, db IDB, obj Model, key string, val any, cols ...string) error {
	return ModelWith(ctx, db, obj, key, "=", val, cols...)
}
func ModelWith(ctx context.Context, db IDB, obj Model, key, op string, val any, cols ...string) error {
	if val == nil || val == 0 || val == "" || op == "" {
		logger().Infow("empty param", "key", key, "op", op, "val", val)
		return ErrEmptyKey
	}
	err := db.NewSelect().Model(obj).Column(cols...).Where("? "+op+" ?", Ident(key), val).Limit(1).Scan(ctx)
	if err == ErrNoRows {
		logger().Debugw("get model with key no rows", "key", key, "op", op, "val", val)
		return ErrNotFound
	}
	if err != nil {
		logger().Warnw("get model with key failed", "key", key, "op", op, "val", val, "err", err)
		return err
	}
	return nil
}

// DoInsert insert with ignore duplicate (optional)
func DoInsert(ctx context.Context, db IDB, obj Model, args ...any) error {
	isZeroID := obj.IsZeroID()
	// Call to saving hook
	if err := TryToBeforeCreateHooks(obj); err != nil {
		return err
	}

	if dtf, ok := obj.(CreatedSetter); ok {
		if ts, ok := CreatedFromContext(ctx); ok && ts > 0 {
			if dtf.SetCreated(ts) {
				logger().Infow("set created ok", "ts", ts)
			} else {
				logger().Infow("set created fail", "ts", ts)
			}
		}
	}

	q := db.NewInsert().Model(obj)
	argc := len(args)
	if argc > 0 {
		unikey := field.ID
		if k, ok := args[0].(string); ok && isZeroID {
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
	q.Returning(field.ID)

	name := GetModelName(q)
	if _, err := q.Exec(ctx); err != nil {
		logger().Infow("insert model fail", "name", name, "obj", obj, "err", err)
		return err
	}

	logger().Debugw("insert model ok", "name", name, "id", obj.GetID())
	if ov, ok := obj.(Changeable); ok && !ov.DisableLog() && operateModelLogFn != nil && isZeroID {
		err := operateModelLogFn(ctx, db, name, OperateTypeCreate, obj)
		if err != nil {
			logger().Infow("call create operateModelLogFn fail", "name", name, "err", err)
		}
	}

	return TryToAfterCreateHooks(obj)
}

func DoUpdate(ctx context.Context, db IDB, obj Model, columns ...string) error {

	// Call to saving hook
	if err := TryToBeforeUpdateHooks(obj); err != nil {
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

	name := GetModelName(q)
	if _, err := q.WherePK().Exec(ctx); err != nil {
		logger().Infow("update model fail", "name", name,
			"obj", obj, "columns", columns, "err", err)
		return err
	}

	if ov, ok := obj.(Changeable); ok && !ov.DisableLog() && operateModelLogFn != nil {
		err := operateModelLogFn(ctx, db, name, OperateTypeUpdate, obj)
		if err != nil {
			logger().Infow("call update operateModelLogFn fail", "name", name, "err", err)
		}
	}
	logger().Debugw("update model ok", "name", name,
		"id", obj.GetID(), "columns", columns)

	if vo, ok := obj.(interface{ SetIsUpdate(v bool) }); ok {
		vo.SetIsUpdate(true)
	}

	return TryToAfterUpdateHooks(obj)
}

func StoreSimple(ctx context.Context, db IDB, obj ModelChangeable, columns ...string) error {
	if !obj.IsZeroID() {
		exist, err := db.NewSelect().Model(obj).WherePK().Column(field.ID).Exists(ctx)
		if err == nil && exist {
			return DoUpdate(ctx, db, obj, columns...)
		}
	}

	return DoInsert(ctx, db, obj)
}

type columnsFn func() []string

// Deprecated: use StoreWithSet[*M]()
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

type ModelSetPtr[T any, U any] interface {
	Model
	*T
	SetWith(in U)
}

// StoreWithSet[*U] save a Model wish ModelSet and value & key
// code examples:
// StoreWithSet[*U](ctx, db, in) // create if no conflict
// StoreWithSet[*U](ctx, db, in, id) // update or create
// StoreWithSet[*U](ctx, db, in, code, "code") // update or create
func StoreWithSet[T ModelSetPtr[U, V], U any, V any](ctx context.Context, db IDB, in V, vk ...string) (obj T, err error) {
	obj = new(U)
	var exist bool
	argc := len(vk)
	if argc > 1 && vk[1] != "" {
		err = ModelWithUnique(ctx, db, obj, vk[1], vk[0])
		exist = (err == nil)
	} else if argc == 1 && obj.SetID(vk[0]) {
		err = ModelWithPK(ctx, db, obj)
		exist = (err == nil)
	}

	obj.SetWith(in)

	DoMetaUp(ctx, db, obj)

	if exist {
		err = DoUpdate(ctx, db, obj)
	} else {
		err = DoInsert(ctx, db, obj)
	}

	return
}

func DoMetaUp(ctx context.Context, db IDB, obj Model) {
	if mm, ok := obj.(ModelMeta); ok {
		for _, f := range metaUpFuncs {
			f(ctx, db, mm)
		}
	}
}

func DoDelete(ctx context.Context, db IDB, table string, _id any) error {
	return DoDeleteT(ctx, db, LastSchema(), LastSchemaCrap(), table, _id)
}

func OpDeleteInTrans(ctx context.Context, db IDB, scDft, scCrap string, tOrQ any, obj any) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx Tx) error {
		var table string
		var name string
		if s, ok := tOrQ.(string); ok {
			table = s
		} else if q, ok := tOrQ.(QueryBase); ok {
			name = GetModelName(q)
			table = GetTableName(q, obj)
		} else {
			panic(fmt.Errorf("invalid %+v", tOrQ))
		}
		var id any
		if v, ok := obj.(Model); ok {
			id = v.GetID()
		} else {
			id = obj
		}
		if err := DoDeleteT(ctx, tx, scDft, scCrap, table, id); err != nil {
			return err
		}
		if ov, ok := obj.(ModelChangeable); ok && !ov.DisableLog() && operateModelLogFn != nil && len(name) > 0 {
			err := operateModelLogFn(ctx, db, name, OperateTypeDelete, ov)
			if err != nil {
				logger().Infow("call delete operateModelLogFn fail", "name", name, "err", err)
			}
		}
		return nil
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

type idsHold struct {
	IDs oid.OIDs `bun:"ids,array"`
}

// BatchDeleteWithKey 按指定的外键批量删除
func BatchDeleteWithKey(ctx context.Context, db IDB, name, key string, id oid.OID) (ids []oid.OID, err error) {
	var hold idsHold
	if err = db.NewRaw("SELECT array_agg(id) as ids FROM ? WHERE ? = ?", Ident(name), Ident(key), id).Scan(ctx, &hold); err == nil {
		ids = hold.IDs
		for _, id := range ids {
			if err = DoDelete(ctx, db, name, id); err != nil {
				logger().Infow("delete fail", "name", name, "key", key, "id", id, "err", err)
				return
			}
		}
		if len(ids) > 0 {
			logger().Infow("batch delete done", "name", name, "key", key, "id", id, "ids", ids)
		}
	} else {
		logger().Infow("query fail when batch delete", "name", name, "key", key, "id", id, "err", err)
	}
	return
}

type Order int8

const (
	OrderAsc Order = iota
	OrderDesc

	OrderNone Order = -1
)

// First find the first model order by pk with condition
//
// Examples:
//
// var user User
// err := First(ctx, db, &User)
//
// var user User
// err := First(ctx, db, &User, id)
//
// var user User
// err := First(ctx, db, &User, "name = ?", "adam")
// ...
func First(ctx context.Context, db IDB, obj Model, args ...any) error {
	return oneWithOrder(ctx, db, OrderAsc, obj, args...)
}

// Get find a model with contition
func Get(ctx context.Context, db IDB, obj Model, arg any, args ...any) error {
	return oneWithOrder(ctx, db, OrderNone, obj, append([]any{arg}, args...)...)
}

// Last find the last model order by pk with condition
func Last(ctx context.Context, db IDB, obj Model, args ...any) error {
	return oneWithOrder(ctx, db, OrderDesc, obj, args...)
}

func oneWithOrder(ctx context.Context, db IDB, ord Order, obj Model, args ...any) error {
	q, ok := QueryOne(db, obj, args...)
	if !ok {
		return ErrInvalidArgs
	}
	switch ord {
	case OrderAsc:
		q.Order("id")
	case OrderDesc:
		q.Order("id DESC")
	}
	err := q.Scan(ctx)
	if err != nil {
		if err == ErrNoRows {
			logger().Debugw("get model with key no rows", "name", GetTableName(q, obj), "args", args)
			return ErrNotFound
		}
		logger().Warnw("get model with key failed", "name", GetTableName(q, obj), "args", args, "err", err)

		if err == ErrBadConn {
			panic(err)
		}
	}
	return err
}

// QueryOne Query one model record base on optional conditions
func QueryOne(db IDB, obj Model, args ...any) (*SelectQuery, bool) {
	q := db.NewSelect().Limit(1).Model(obj)

	if len(args) == 0 {
		if obj.IsZeroID() {
			// unconditional
			return q, true
		}
		return q.WherePK(), true
	}
	if len(args) == 1 {
		obj.SetID(args[0])
		return q.WherePK(), true
	}
	if s, ok := args[0].(string); ok {
		return q.Where(s, args[1:]...), true
	}

	logger().Infow("queryOne: invalid args", "name", GetTableName(q, obj), "args", args)

	return q, false
}

// QueryList Query as a collection list with a Sifter
func QueryList(ctx context.Context, db IDB, spec Sifter, dataptr any) *SelectQuery {
	q := db.NewSelect().Model(dataptr)
	if v, ok := spec.(SifterX); ok {
		q = v.SiftX(ctx, q)
	}
	if !spec.IsSifted() {
		q = spec.Sift(q)
	}

	return q
}

// ApplyQueryContext Apply column filtering in a query by reading the context
func ApplyQueryContext(ctx context.Context, q *SelectQuery) *SelectQuery {
	if excols := ExcludesFromContext(ctx); len(excols) > 0 {
		q.ExcludeColumn(excols...)
	} else if cols := ColumnsFromContext(ctx); len(cols) > 0 {
		q.Column(cols...)
	}
	return q
}
