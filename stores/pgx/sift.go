package pgx

import (
	"time"

	"daxv.cn/gopak/lib/sqlutil"
	"hyyl.xyz/cupola/andvari/models/oid"
	"hyyl.xyz/cupola/andvari/utils"
)

// MDftSpec 默认的查询条件
type MDftSpec struct {
	// 主键编号`ids`（以逗号分隔的字串），仅供 Form 或 Query 使用, example:"aaa,bbb,ccc"
	IDsStr oid.OIDsStr `form:"ids" json:"idstr"  extensions:"x-order=0" example:"aaa,bbb,ccc"`
	// 主键编号`ids`（集），仅供 JSON 使用, example:"['aaa','bbb','ccc']"
	IDs oid.OIDs `form:"-" json:"ids"  extensions:"x-order=1" `
	// 创建者ID
	CreatorID string `form:"creatorID" json:"creatorID"  extensions:"x-order=2"`
	// 创建时间 形式： yyyy-mm-dd, 1_day, 2_weeks, 3_months
	Created string `form:"created" json:"created"  extensions:"x-order=3"`
	// 更新时间 形式： yyyy-mm-dd, 1_day, 2_weeks, 3_months
	Updated string `form:"updated" json:"updated"  extensions:"x-order=4"`
} // @name DefaultSpec

// CanSort 检测字段是否可排序
func (md *MDftSpec) CanSort(key string) bool {
	switch key {
	case "id", "created", "updated":
		return true
	default:
		return false
	}
}

func (md *MDftSpec) Sift(q *ormQuery) (*ormQuery, error) {
	tm := q.TableModel()
	var pre string
	if len(tm.GetJoins()) > 0 {
		pre = string(tm.Table().Alias) + "."
	}
	if len(md.IDs) > 0 {
		q.WhereIn(pre+"id in (?)", md.IDs)
	} else if md.IDsStr.Valid() {
		ids, err := md.IDsStr.Decode()
		if err != nil {
			return nil, err
		}
		q.WhereIn(pre+"id in (?)", ids)
	}
	q, _ = SiftOID(q, "creator_id", md.CreatorID, false)
	q, _ = SiftDate(q, "created", md.Created, false, false)
	q, _ = SiftDate(q, "updated", md.Updated, false, false)

	return q, nil
}

func SiftOID(q *ormQuery, field string, s string, isOr bool) (*ormQuery, bool) {
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
func SiftEquel(q *ormQuery, field string, v any, isOr bool) (*ormQuery, bool) {
	return Sift(q, field, "=", v, isOr)
}

// SiftICE 忽略大小写相等
func SiftICE(q *ormQuery, field string, v string, isOr bool) (*ormQuery, bool) {
	if utils.IsZero(v) {
		return q, false
	}
	return Sift(q, field, "ILIKE", sqlutil.CleanWildcard(v), isOr)
}

// SiftMatch 忽略大小写并匹配前缀
func SiftMatch(q *ormQuery, field string, v string, isOr bool) (*ormQuery, bool) {
	if utils.IsZero(v) {
		return q, false
	}
	return Sift(q, field, "ILIKE", sqlutil.MendValue(v), isOr)
}

var SiftILike = SiftMatch // Deprecated

// SiftGreat 大于
func SiftGreat(q *ormQuery, field string, v any, isOr bool) (*ormQuery, bool) {
	return Sift(q, field, ">", v, isOr)
}

// SiftLess 小于
func SiftLess(q *ormQuery, field string, v any, isOr bool) (*ormQuery, bool) {
	return Sift(q, field, "<", v, isOr)
}

func Sift(q *ormQuery, field, op string, v any, isOr bool) (*ormQuery, bool) {
	tm := q.TableModel()
	var pre string
	if len(tm.GetJoins()) > 0 {
		pre = string(tm.Table().Alias) + "."
	}
	if t, ok := v.(time.Time); ok {
		if t.IsZero() {
			return q, false
		}
		if op == "=" {
			const oneDay = time.Hour * 24
			if isOr {
				return q.WhereOr(pre+"? BETWEEN ? AND ?", pgIdent(field),
					t.Truncate(oneDay), t.Add(oneDay).Truncate(oneDay)), true
			}
			return q.Where(pre+"? BETWEEN ? AND ?", pgIdent(field),
				t.Truncate(oneDay), t.Add(oneDay).Truncate(oneDay)), true
		}
	}
	if utils.IsZero(v) {
		return q, false
	}
	if isOr {
		return q.WhereOr(pre+"? "+op+" ?", pgIdent(field), v), true
	}
	return q.Where(pre+"? "+op+" ?", pgIdent(field), v), true
}

// SiftDate 按日期(时间)类型传递查询条件
//
//	during 符合 GetDateRange 参数格式
//	isInt 是指用整数(毫秒)表示的时间
func SiftDate(q *ormQuery, field string, during string, isInt, isOr bool) (*ormQuery, bool) {
	if len(during) > 0 {
		tm := q.TableModel()
		var pre string
		if len(tm.GetJoins()) > 0 {
			pre = string(tm.Table().Alias) + "."
		}
		if dr, err := sqlutil.GetDateRange(during); err == nil {
			var start, end any
			if isInt {
				start, end = dr.Start.UnixMilli(), dr.End.UnixMilli()
			} else {
				start, end = dr.Start, dr.End
			}
			if isOr {
				return q.WhereOr(pre+"? between ? and ?", pgIdent(field), start, end), true
			}
			return q.Where(pre+"? between ? and ?", pgIdent(field), start, end), true
		} else {
			logger().Infow("invalid param", "field", field, "during", during, "err", err)
		}
	}

	return q, false
}
