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
	// 搜索风格 `web` `plain` 或空
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

func (tss *TextSearchSpec) Sift(q *SelectQuery) (*SelectQuery, error) {
	return DoApplyTsQuery(tss.enabled, tss.cfgname, q, tss.SearchKeyWord, tss.SearchStyle, tss.fallbacks...)
}

func DoApplyTsQuery(enabled bool, cfgname string, q *SelectQuery, kw, sty string, cols ...string) (*SelectQuery, error) {
	if len(kw) == 0 {
		return q, nil
	}
	if enabled {
		return q.Where("? @@ "+getTsQuery(cfgname, sty, kw), Ident(textVec)), nil
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
	return q, nil
}

func getTsQuery(tscfg string, sty, kw string) string {
	return fmt.Sprintf("%s('%s', '%s')", GetTSQname(sty), tscfg, sqlutil.CleanWildcard(kw))
}

func CheckTsCfg(ctx context.Context, db IDB, ftsConfig string) bool {
	var ret int
	err := db.NewSelect().Table("pg_ts_config").Column("oid").Where("cfgname = ?", ftsConfig).Scan(ctx, &ret)
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
