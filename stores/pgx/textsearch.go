package pgx

import (
	"context"
	"fmt"

	"github.com/cupogo/andvari/models/field"
	"github.com/cupogo/andvari/utils/sqlutil"
)

const (
	textVec = "ts_vec"
)

type TextSearchSpec struct {
	cfgname string
	enabled bool

	fallbacks []string // columns
	bothmatch bool     // both left and right match '%abc%'

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

// Deprecated: use SetTsFallback
func (tss *TextSearchSpec) SetFallback(cols ...string) {
	tss.fallbacks = cols
}

func (tss *TextSearchSpec) SetBothMatch(yn bool) {
	tss.bothmatch = yn
}

func (tss *TextSearchSpec) TsEnabled() bool {
	return tss.enabled
}

func HideTsColumn(q *SelectQuery) *SelectQuery {
	return q.ExcludeColumn(field.TsCfg, field.TsVec)
}

func (tss *TextSearchSpec) Sift(q *SelectQuery) *SelectQuery {
	return tss.SiftTS(q, true)
}
func (tss *TextSearchSpec) SiftTS(q *SelectQuery, hide bool) *SelectQuery {
	if tss.enabled && hide {
		q = HideTsColumn(q)
	}
	if len(tss.SearchKeyWord) == 0 {
		return q
	}
	return q.WhereGroup(" AND ", func(sq *SelectQuery) *SelectQuery {
		if len(tss.fallbacks) > 0 && len(tss.fallbacks[0]) > 0 {
			for _, col := range tss.fallbacks {
				sq.WhereOr("? iLIKE ?", Ident(col), sqlutil.MendValue(tss.SearchKeyWord, tss.bothmatch))
			}
		}
		if tss.enabled {
			sq.WhereOr("? @@ "+getTsQuery(tss.cfgname, tss.SearchStyle, tss.SearchKeyWord), Ident(textVec))
		}
		return sq
	})
}

// Deprecated: use TextSearchSpec.Sift
func DoApplyTsQuery(enabled bool, cfgname string, q *SelectQuery, kw, sty string, cols ...string) *SelectQuery {
	tss := TextSearchSpec{
		cfgname:       cfgname,
		enabled:       enabled,
		SearchKeyWord: kw,
		SearchStyle:   sty,
		fallbacks:     cols,
	}
	return tss.Sift(q)
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
		logger().Infow("fts checked fail", "tscfg", ftsConfig, "err", err)
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
