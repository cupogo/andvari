package pgx

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/cupogo/andvari/models/comm"
	"github.com/cupogo/andvari/models/oid"
	"github.com/cupogo/andvari/utils/zlog"
)

func TestMain(m *testing.M) {
	lgr, _ := zap.NewDevelopment()
	defer func() {
		_ = lgr.Sync() // flushes buffer, if any
	}()
	sugar := lgr.Sugar()
	zlog.Set(sugar)

	ret := m.Run()

	os.Exit(ret)
}

// CREATE ROLE testing WITH LOGIN PASSWORD 'testing1';
// CREATE DATABASE testing WITH OWNER = testing ENCODING = 'UTF8';
// GRANT ALL PRIVILEGES ON DATABASE testing to testing;
func TestOpen(t *testing.T) {
	dsn := "postgres://testing@localhost"
	ftscfg := ""
	db, err := Open(dsn, ftscfg, false)
	assert.Error(t, err)
	assert.Nil(t, db)

	dsn = "postgres://testing:testing1@localhost/postgres?sslmode=disable"
	db, err = Open(dsn, ftscfg, false)
	assert.Error(t, err)
	assert.Nil(t, db)

	dsn = "postgres://testing:testing1@localhost/testing?sslmode=disable"
	db, err = Open(dsn, ftscfg, false)
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

// Clause 条款
type Clause struct {
	comm.BaseModel `bun:"table:cms_clause,alias:c" json:"-"`

	comm.DefaultModel

	ClauseBasic
} // @name Clause

type ClauseBasic struct {
	Text string `bun:"text,notnull,type:text" extensions:"x-order=A" form:"text" json:"text" pg:"text,notnull"`
} // @name ClauseBasic

type Clauses []Clause

// Creating function call to it's inner fields defined hooks
func (z *Clause) Creating() error {
	if z.IsZeroID() {
		z.SetID(oid.NewID(oid.OtArticle))
	}

	return z.DefaultModel.Creating()
}

type ClauseSpec struct {
	comm.PageSpec
	ModelSpec

	Text string `extensions:"x-order=A" form:"text" json:"text"`
}

func (spec *ClauseSpec) Sift(q *SelectQuery) *SelectQuery {
	q = spec.ModelSpec.Sift(q)
	q, _ = SiftMatch(q, "text", spec.Text, false)

	return q
}

func TestOps(t *testing.T) {
	db, err := Open("postgres://testing:testing1@localhost/testing?sslmode=disable", "", true)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	ctx := context.Background()
	dropIt := true
	tables := []any{&Clause{}}
	err = CreateModels(ctx, db, dropIt, tables...)
	assert.NoError(t, err)

	obj := new(Clause)
	obj.Text = "test"
	err = DoInsert(ctx, db, obj)
	assert.NoError(t, err)
	assert.False(t, obj.IsZeroID())
	err = DoInsert(ctx, db, obj, "text")
	assert.NoError(t, err)

	obj2 := new(Clause)
	obj2.Text = "hello world"
	err = StoreSimple(ctx, db, obj2, "text")
	assert.NoError(t, err)
	assert.False(t, obj.IsZeroID())
	err = StoreSimple(ctx, db, obj2, "text")
	assert.NoError(t, err)

	spec := &ClauseSpec{}
	spec.Limit = 2
	spec.Text = "test"
	var data Clauses
	q := db.NewSelect().Model(&data)
	total, err := QueryPager(ctx, spec, q.Apply(spec.Sift))
	assert.NoError(t, err)
	assert.NotZero(t, total)

	exist := new(Clause)
	err = ModelWithPKID(ctx, db, exist, obj.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test", exist.Text)
	exist.Text = "test2"
	err = DoUpdate(ctx, db, exist, "text")
	assert.NoError(t, err)

}
