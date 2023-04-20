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

const (
	testDSN = "postgres://testing:testing1@localhost/testing?sslmode=disable"
)

func TestInit(t *testing.T) {
	db, err := Open(testDSN, "simple", false)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	ctx := context.Background()
	dropIt := true
	err = db.InitSchemas(ctx, dropIt)
	assert.NoError(t, err)

	err = db.RunMigrations(ctx, fstest.MapFS{})
	assert.NoError(t, err)

	var opts []AlterOption
	opts = append(opts, WithAlterAdd())
	opts = append(opts, WithAlterDrop())
	err = db.SyncSchema(ctx, opts...)
	assert.NoError(t, err)

	ListFS("init", os.Stderr)
	ListFS("alter", os.Stderr)
}

// Clause 条款
type Clause struct {
	comm.BaseModel `bun:"table:cms_clause,alias:c" json:"-"`

	comm.DefaultModel

	ClauseBasic
	comm.MetaField
} // @name Clause

type ClauseBasic struct {
	Slug  string   `bun:"slug,notnull,type:name,unique" json:"slug" pg:"slug,notnull"`
	Text  string   `bun:"text,notnull,type:text" form:"text" json:"text" pg:"text,notnull"`
	Cates []string `bun:"cates,notnull,type:jsonb" form:"cats" json:"cats" pg:"cates,notnull,type:jsonb"`
} // @name ClauseBasic

type Clauses []Clause

// Creating function call to it's inner fields defined hooks
func (z *Clause) Creating() error {
	if z.IsZeroID() {
		z.SetID(oid.NewID(oid.OtArticle))
	}

	return z.DefaultModel.Creating()
}

type ClauseSet struct {
	Slug *string `extensions:"x-order=A" json:"slug"`
	Text *string `extensions:"x-order=B" json:"text"`
} // @name cms1ClauseSet

func (z *Clause) SetWith(o ClauseSet) {
	if o.Slug != nil && z.Slug != *o.Slug {
		z.LogChangeValue("slug", z.Slug, o.Slug)
		z.Slug = *o.Slug
	}
	if o.Text != nil && z.Text != *o.Text {
		z.LogChangeValue("text", z.Text, o.Text)
		z.Text = *o.Text
	}
}

func dbOpModelMeta(ctx context.Context, db IDB, obj Model) error { return nil }

func init() {
	RegisterModel((*Clause)(nil))
	RegisterDbFs(embeds.DBFS())
	RegisterMetaUp(dbOpModelMeta)
}

type ClauseSpec struct {
	comm.PageSpec
	ModelSpec

	Text  string   `form:"text" json:"text"`
	Cates []string `form:"cats" json:"cats"`
}

func (spec *ClauseSpec) Sift(q *SelectQuery) *SelectQuery {
	q = spec.ModelSpec.Sift(q)
	q, _ = SiftMatch(q, "text", spec.Text, false)
	q, _ = Sift(q, "cates", "any", spec.Cates, false)

	return q
}

func TestOps(t *testing.T) {
	db, err := Open(testDSN, "mycfg", true)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	ctx := context.Background()

	err = DoInsert(ctx, db, &Clause{})
	assert.NoError(t, err)

	obj := new(Clause)
	obj.Slug = oid.NewObjID(oid.OtDefault)
	obj.Text = "test"
	obj.Cates = append(obj.Cates, "cat", "dog")
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

	nid := obj.ID
	one := new(Clause)
	assert.NoError(t, First(ctx, db, one))
	assert.NotZero(t, one.ID)
	one = new(Clause)
	assert.NoError(t, First(ctx, db, one, nid))
	assert.NotZero(t, one.ID)
	one = new(Clause)
	assert.NoError(t, First(ctx, db, one, nid.String()))
	assert.NotZero(t, one.ID)
	one = new(Clause)
	assert.NoError(t, First(ctx, db, one, "id = ?", nid))
	assert.NotZero(t, one.ID)
	one = new(Clause)
	assert.NoError(t, First(ctx, db, one, "slug = ?", obj.Slug))
	assert.NotZero(t, one.ID)
	one = new(Clause)
	assert.NoError(t, Get(ctx, db, one, nid))
	assert.NotZero(t, one.ID)
	one = new(Clause)
	assert.NoError(t, Last(ctx, db, one))
	assert.NotZero(t, one.ID)

	one = new(Clause)
	assert.Error(t, Last(ctx, db, one, 0))
	assert.Zero(t, one.ID)
	one = new(Clause)
	assert.Error(t, Last(ctx, db, one, ""))
	assert.Zero(t, one.ID)
	one = new(Clause)
	assert.Error(t, Last(ctx, db, one, 1, 2))

	spec := &ClauseSpec{}
	spec.Limit = 2
	spec.Cates = append(spec.Cates, "dog", "sheep")
	spec.Text = "test"
	spec.Sort = "created DESC"
	spec.Created = "0_day"
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
	assert.True(t, exist.IsUpdate())
	exist.Text = "test2"
	err = DoUpdate(ctx, db, exist, "text")
	assert.NoError(t, err)

	exist = new(Clause)
	err = db.GetModel(ctx, exist, "")
	assert.Error(t, err)
	err = db.GetModel(ctx, exist, "not-found")
	assert.Error(t, err)
	err = db.GetModel(ContextWithColumns(ctx, "text"), exist, obj.ID)
	assert.NoError(t, err)
	err = db.GetModel(ContextWithExcludes(ctx, "text"), exist, obj.ID)
	assert.NoError(t, err)
	err = db.GetModel(ctx, exist, obj.ID, "text")
	assert.NoError(t, err)

	slug := "eagle"
	text := "hawk"
	in := ClauseSet{Slug: &slug, Text: &text}
	obj3, err := StoreWithSet[*Clause](ctx, db, in, slug, "slug")
	t.Logf("obj: %+v", obj3)
	assert.NoError(t, err)
	assert.NotNil(t, obj3)
	assert.NotZero(t, obj3.ID)
	assert.Equal(t, text, obj3.Text)
	text = "kite"
	in.Text = &text
	obj4, err := StoreWithSet[*Clause](ctx, db, in, obj3.StringID())
	assert.NoError(t, err)
	assert.NotNil(t, obj4)
	assert.NotZero(t, obj4.ID)
	assert.Equal(t, text, obj4.Text)
	assert.Equal(t, obj3.ID, obj4.ID)
	assert.Equal(t, obj3.Slug, obj4.Slug)

	_, err = StoreWithSet[*Clause](ctx, db, in)
	assert.NoError(t, err)
	_, err = StoreWithSet[*Clause](ctx, db, in, "")
	assert.NoError(t, err)
	_, err = StoreWithSet[*Clause](ctx, db, in, "", "slug")
	assert.Error(t, err)
}
