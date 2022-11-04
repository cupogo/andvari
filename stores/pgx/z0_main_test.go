package pgx

import (
	"context"
	"os"
	"testing"
	"testing/fstest"
	"time"

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
	db, err := Open("postgres://testing:testing1@localhost/testing?sslmode=disable", "simple", false)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	ctx := context.Background()
	dropIt := true
	err = db.InitSchemas(ctx, dropIt)
	assert.NoError(t, err)

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
	Slug string `bun:"slug,notnull,type:name,unique" extensions:"x-order=A" form:"slug" json:"slug" pg:"slug,notnull"`
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
	db, err := Open("postgres://testing:testing1@localhost/testing?sslmode=disable", "mycfg", true)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	ctx := context.Background()

	err = DoInsert(ctx, db, &Clause{})
	assert.NoError(t, err)

	obj := new(Clause)
	obj.Slug = oid.NewObjID(oid.OtDefault)
	obj.Text = "test"
	err = DoInsert(ContextWithCreated(ctx, 0), db, obj, "slug")
	assert.NoError(t, err)
	assert.False(t, obj.IsZeroID())
	assert.NotZero(t, obj.ID)
	err = DoInsert(ctx, db, obj, "slug", "text")
	assert.NoError(t, err)
	assert.False(t, obj.IsZeroID())
	assert.NotZero(t, obj.ID)

	now := time.Now()
	obj2 := new(Clause)
	obj2.Slug = oid.NewObjID(oid.OtDefault)
	obj2.Text = "hello world"
	err = StoreSimple(ContextWithCreated(ctx, now.UnixMilli()), db, obj2, "text")
	assert.NoError(t, err)
	assert.False(t, obj2.IsZeroID())
	assert.NotZero(t, obj2.ID)
	assert.False(t, obj.IsZeroID())
	err = StoreSimple(ctx, db, obj2, "text")
	assert.NoError(t, err)
	assert.False(t, obj2.IsZeroID())
	assert.NotZero(t, obj2.ID)

	spec := &ClauseSpec{}
	spec.Limit = 2
	spec.Text = "test"
	spec.Sort = "created DESC"
	var data Clauses
	total, err := db.List(ContextWithColumns(ctx, "text"), spec, &data)
	assert.NoError(t, err)
	assert.NotZero(t, total)

	id := obj2.ID
	err = db.DeleteModel(ctx, &Clause{}, id)
	assert.NoError(t, err)

	err = db.GetModel(ctx, &Clause{}, id)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrNotFound.Error())

	spec2 := &ClauseSpec{}
	spec2.Limit = 2
	spec2.IsDelete = true
	assert.True(t, spec2.Deleted())
	var data2 Clauses
	total, err = db.List(ContextWithExcludes(ctx, "text"), spec2, &data2)
	assert.NoError(t, err)
	assert.NotZero(t, total)
	if assert.NotEmpty(t, data2) {
		assert.Empty(t, data2[0].Text)
	}

	err = db.UndeleteModel(ctx, &Clause{}, id)
	assert.NoError(t, err)
	err = db.GetModel(ctx, &Clause{}, id)
	assert.NoError(t, err)

	exist := new(Clause)
	err = ModelWithPKID(ctx, db, exist, obj.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test", exist.Text)
	exist.Text = ""
	err = DoUpdate(ctx, db, exist, "text")
	assert.NoError(t, err)
	exist.Text = "test2"
	err = DoUpdate(ctx, db, exist, "text")
	assert.NoError(t, err)

	exist = new(Clause)
	err = db.GetModel(ctx, exist, "")
	assert.NoError(t, err)
	err = db.GetModel(ctx, exist, "not-found")
	assert.Error(t, err)
	err = db.GetModel(ContextWithColumns(ctx, "text"), exist, obj.ID)
	assert.NoError(t, err)
	err = db.GetModel(ContextWithExcludes(ctx, "text"), exist, obj.ID)
	assert.NoError(t, err)
	err = db.GetModel(ctx, exist, obj.ID, "text")
	assert.NoError(t, err)
}
