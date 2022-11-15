package pgx

import (
	"context"
	"strings"
	"time"

	"github.com/cupogo/andvari/models/oid"
	"github.com/cupogo/andvari/utils"
	"github.com/cupogo/andvari/utils/sqlutil"
)

type Sifter interface {
	Sift(q *SelectQuery) *SelectQuery
}

type SifterX interface {
	SiftX(ctx context.Context, q *SelectQuery) *SelectQuery
}

type ListArg interface {
	Pager
	Sifter
	Deleted() bool
}

// StringsDiff 字串增减操作
type StringsDiff struct {
	Newest  []string `json:"newest" validate:"dive"`  // 新增的字串集
	Removed []string `json:"removed" validate:"dive"` // 删除的字串集
} // @name StringsDiff

// ModelSpec 模型默认的查询条件
type ModelSpec struct {
	// 主键编号`ids`（以逗号分隔的字串），仅供 Form 或 Query 使用, example:"aaa,bbb,ccc"
	IDsStr oid.OIDsStr `form:"ids" json:"ids"  extensions:"x-order=0" example:"aaa,bbb,ccc"`
	// 主键编号`ids`（集），仅供 JSON 使用, example:"['aaa','bbb','ccc']"
	IDs oid.OIDs `form:"-" json:"idArr"   swaggerignore:"true"`
	// 创建者ID
	CreatorID string `form:"creatorID" json:"creatorID"  extensions:"x-order=2"`
	// 创建时间 形式： yyyy-mm-dd, 1_day, 2_weeks, 3_months
	Created string `form:"created" json:"created"  extensions:"x-order=3"`
	// 更新时间 形式： yyyy-mm-dd, 1_day, 2_weeks, 3_months
	Updated string `form:"updated" json:"updated"  extensions:"x-order=4"`
	// IsDelete 查询删除的记录
	IsDelete bool `form:"isDelete" json:"isDelete"  extensions:"x-order=5"`
} // @name DefaultSpec

// CanSort 检测字段是否可排序
func (md *ModelSpec) CanSort(key string) bool {
	switch key {
	case "id", "created", "updated":
		return true
	default:
		return false
	}
}

func (md *ModelSpec) Deleted() bool {
	return md.IsDelete
}

func (md *ModelSpec) Sift(q *SelectQuery) *SelectQuery {
	if len(md.IDs) > 0 {
		q.Where("?TableAlias.id in (?)", In(md.IDs))
	} else if md.IDsStr.Valid() {
		ids, err := md.IDsStr.Decode()
		if err == nil {
			q.Where("?TableAlias.id in (?)", In(ids))
		}
	}
	q, _ = SiftOID(q, "creator_id", md.CreatorID, false)
	q, _ = SiftDate(q, "created", md.Created, false, false)
	q, _ = SiftDate(q, "updated", md.Updated, false, false)

	return q
}

func SiftOIDs(q *SelectQuery, field string, s string, isOr bool) (*SelectQuery, bool) {
	if len(s) > 0 {
		if ids, ok := oid.ParseOIDs(s); ok {
			if len(ids) == 1 {
				return Sift(q, field, "=", ids[0], isOr)
			}
			return Sift(q, field, "in", ids, isOr)
		} else {
			logger().Infow("invalid oids", "s", s)
		}
	}
	return q, false
}

func SiftOID(q *SelectQuery, field string, s string, isOr bool) (*SelectQuery, bool) {
	if len(s) > 0 {
		if _, id, err := oid.Parse(s); err == nil {
			return Sift(q, field, "=", id, isOr)
		} else {
			logger().Infow("invalid oid", "s", s)
		}
	}
	return q, false
}

// SiftEquel 完全相等
func SiftEquel(q *SelectQuery, field string, v any, isOr bool) (*SelectQuery, bool) {
	return Sift(q, field, "=", v, isOr)
}

// SiftICE ignore case equel 忽略大小写相等
func SiftICE(q *SelectQuery, field string, v string, isOr bool) (*SelectQuery, bool) {
	if utils.IsZero(v) {
		return q, false
	}
	return Sift(q, field, "ILIKE", sqlutil.CleanWildcard(v), isOr)
}

// SiftMatch ignore case match 忽略大小写并匹配前缀
func SiftMatch(q *SelectQuery, field string, v string, isOr bool) (*SelectQuery, bool) {
	if utils.IsZero(v) {
		return q, false
	}
	return Sift(q, field, "ILIKE", sqlutil.MendValue(v), isOr)
}

// SiftGreat 大于
func SiftGreat(q *SelectQuery, field string, v any, isOr bool) (*SelectQuery, bool) {
	return Sift(q, field, ">", v, isOr)
}

// SiftLess 小于
func SiftLess(q *SelectQuery, field string, v any, isOr bool) (*SelectQuery, bool) {
	return Sift(q, field, "<", v, isOr)
}

func Sift(q *SelectQuery, field, op string, v any, isOr bool) (*SelectQuery, bool) {
	if utils.IsZero(v) || len(field) == 0 {
		return q, false
	}

	if t, ok := v.(time.Time); ok {
		if t.IsZero() {
			return q, false
		}
		if op == "=" {
			const oneDay = time.Hour * 24
			return SiftBetween(q, field, t.Truncate(oneDay), t.Add(oneDay).Truncate(oneDay), isOr)
		}
	}

	var cond string
	if strings.ToUpper(op) == "IN" {
		cond = "? " + op + " (?)"
		if _, ok := v.(QueryAppender); !ok {
			v = In(v)
		}
	} else {
		if op == "?|" || strings.ToUpper(op) == "ANY" {
			op = "\\?|"
			v = Array(v)
		}
		cond = "? " + op + " ?"
	}
	if !strings.Contains(field, ".") {
		cond = "?TableAlias." + cond
	}
	if isOr {
		return q.WhereOr(cond, Ident(field), v), true
	}
	return q.Where(cond, Ident(field), v), true
}

// SiftDate 按日期(时间)类型传递查询条件
//
//	during 符合 GetDateRange 参数格式
//	isInt 是指用整数(毫秒)表示的时间
func SiftDate(q *SelectQuery, field string, during string, isInt, isOr bool) (*SelectQuery, bool) {
	if len(during) > 0 {
		if dr, err := sqlutil.GetDateRange(during); err == nil {
			var start, end any
			if isInt {
				start, end = dr.Start.UnixMilli(), dr.End.UnixMilli()
			} else {
				start, end = dr.Start, dr.End
			}
			return SiftBetween(q, field, start, end, isOr)
		} else {
			logger().Infow("invalid param", "field", field, "during", during, "err", err)
		}
	}

	return q, false
}

// SiftBetween 匹配两个值之间的条件
func SiftBetween(q *SelectQuery, field string, v1, v2 any, isOr bool) (*SelectQuery, bool) {
	if utils.IsZero(v1) || utils.IsZero(v2) {
		return q, false
	}
	cond := "? "
	if !strings.Contains(field, ".") {
		cond = "?TableAlias." + cond
	}
	if isOr {
		return q.WhereOr(cond+" BETWEEN ? AND ?", Ident(field), v1, v2), true
	}
	return q.Where(cond+" BETWEEN ? AND ?", Ident(field), v1, v2), true

}
