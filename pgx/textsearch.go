package pgx

import (
	"fmt"

	"daxv.cn/gopak/lib/sqlutil"
)

const (
	textVec = "ts_vec"
)

type TextSearchSpec struct {
	cfgname string
	enabled bool

	fallbacks []string // columns

	// 关键词搜索
	SearchKeyWord string `json:"skw,omitempty" form:"skw" extensions:"x-order=8"`
	// 搜索风格 `web` `plain` 或空
	SearchStyle string `json:"sst,omitempty" form:"sst" extensions:"x-order=9"`
}

func (tss *TextSearchSpec) SetFallback(cols ...string) {
	tss.fallbacks = cols
}

func (tss *TextSearchSpec) Sift(q *ormQuery) (*ormQuery, error) {
	return DoApplyTsQuery(tss.enabled, tss.cfgname, q, tss.SearchKeyWord, tss.SearchStyle, tss.fallbacks...)
}

func DoApplyTsQuery(enabled bool, cfgname string, q *ormQuery, kw, sty string, cols ...string) (*ormQuery, error) {
	if len(kw) == 0 {
		return q, nil
	}
	if enabled {
		return q.Where("? @@ "+getTsQuery(cfgname, sty, kw), pgIdent(textVec)), nil
	}
	if len(cols) > 0 && len(cols[0]) > 0 {
		q.WhereGroup(func(_q *ormQuery) (*ormQuery, error) {
			for i, col := range cols {
				if i == 0 {
					_q.Where("? iLIKE ?", pgIdent(col), sqlutil.MendValue(kw))
				} else {
					_q.WhereOr("? iLIKE ?", pgIdent(col), sqlutil.MendValue(kw))
				}
			}
			return _q, nil
		})
	}
	return q, nil
}

func getTsQuery(tscfg string, sty, kw string) string {
	return fmt.Sprintf("%s('%s', '%s')", GetTSQname(sty), tscfg, sqlutil.CleanWildcard(kw))
}

// GetTSQname return ts func name with search style
func GetTSQname(sty string) string {
	switch sty {
	case "web":
		return "websearch_to_tsquery"
	case "plain":
		return "plainto_tsquery"
	default:
		return "to_tsquery"
	}
}
