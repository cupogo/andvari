package pgx

import (
	"context"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/cupogo/andvari/database/embeds"
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

}

func TestInit(t *testing.T) {
	db, err := Open("postgres://testing:testing1@localhost/testing?sslmode=disable", "", false)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	ctx := context.Background()
	dropIt := true
	err = db.InitSchemas(ctx, dropIt)

	err = db.RunMigrations(ctx, fstest.MapFS{})
	assert.NoError(t, err)
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

func init() {
	RegisterModel((*Clause)(nil))
	RegisterDbFs(embeds.DBFS())
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
	total, err := db.List(ContextWithColumns(ctx, "text"), spec, &data)
	assert.NoError(t, err)
	assert.NotZero(t, total)

	err = db.OpDeleteOID(ctx, "cms_clause", obj2.StringID())
	assert.NoError(t, err)

	spec2 := &ClauseSpec{}
	spec2.Limit = 2
	spec2.IsDelete = true
	assert.True(t, spec2.Deleted())
	var data2 Clauses
	total, err = db.List(ContextWithExcludes(ctx, "text"), spec2, &data2)
	assert.NoError(t, err)
	assert.NotZero(t, total)

	err = db.OpUndeleteOID(ctx, "cms_clause", obj2.StringID())
	assert.NoError(t, err)

	exist := new(Clause)
	err = ModelWithPKID(ctx, db, exist, obj.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test", exist.Text)
	exist.Text = "test2"
	err = DoUpdate(ctx, db, exist, "text")
	assert.NoError(t, err)

}
