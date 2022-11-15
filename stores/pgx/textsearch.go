package pgx

import (
	"context"
	"fmt"

	"github.com/cupogo/andvari/utils/sqlutil"
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
	// 匹配风格 `web` `plain` `valid` 或空
	SearchStyle string `json:"sst,omitempty" form:"sst" extensions:"x-order=9" enums:",web,plain"`
}

func (tss *TextSearchSpec) SetTsConfig(cn string, en bool) {
	tss.cfgname, tss.enabled = cn, en
}

func (tss *TextSearchSpec) SetTsFallback(cols ...string) {
	tss.fallbacks = cols
}

// deprecated
func (tss *TextSearchSpec) SetFallback(cols ...string) {
	tss.fallbacks = cols
}

func (tss *TextSearchSpec) Sift(q *SelectQuery) *SelectQuery {
	return DoApplyTsQuery(tss.enabled, tss.cfgname, q, tss.SearchKeyWord, tss.SearchStyle, tss.fallbacks...)
}

func DoApplyTsQuery(enabled bool, cfgname string, q *SelectQuery, kw, sty string, cols ...string) *SelectQuery {
	if len(kw) == 0 {
		return q
	}
	if enabled {
		return q.Where("? @@ "+getTsQuery(cfgname, sty, kw), Ident(textVec))
	}
	if len(cols) > 0 && len(cols[0]) > 0 {
		q.WhereGroup(" AND ", func(_q *SelectQuery) *SelectQuery {
			for i, col := range cols {
				if i == 0 {
					_q.Where("? iLIKE ?", Ident(col), sqlutil.MendValue(kw))
				} else {
					_q.WhereOr("? iLIKE ?", Ident(col), sqlutil.MendValue(kw))
				}
			}
			return _q
		})
	}
	return q
}

func getTsQuery(tscfg string, sty, kw string) string {
	return fmt.Sprintf("%s('%s', '%s')", GetTSQname(sty), tscfg, sqlutil.CleanWildcard(kw))
}

func CheckTsCfg(ctx context.Context, db IDB, ftsConfig string) bool {
	if len(ftsConfig) == 0 {
		return false
	}
	var ret int
	err := db.NewSelect().Table("pg_ts_config").Column("oid").Where("cfgname = ?", ftsConfig).Scan(ctx, &ret)
	if err == nil {
		if ret > 0 {
			logger().Debugw("fts checked ok", "ts cfg", ftsConfig)
			return true
		}
	} else {
		logger().Infow("select ts config fail", "tscfg", ftsConfig, "err", err)
	}
	return false
}

// GetTSQname return ts func name with search style
// see also https://www.postgresql.org/docs/13/functions-textsearch.html
func GetTSQname(sty string) string {
	switch sty {
	case "web":
		return "websearch_to_tsquery"
	case "plain":
		return "plainto_tsquery"
	case "valid":
		return "to_tsquery"
	default:
		return "phraseto_tsquery"
	}
}
