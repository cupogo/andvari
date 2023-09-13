# Andvari

基础模型和数据访问组件

Base model and data access components

Dependencies
---

- [Bun](https://bun.uptrace.dev/)

Databases
---

- `PostgreSQL` for now


Interfaces
---
```go

// Changeable can update a model with special columns
type Changeable interface {
	SetChange(...string)
	GetChanges() []string
	CountChange() int
}

// Model based primary key ID
type Model interface {
	GetID() any
	SetID(id any) bool
	IsZeroID() bool
}

type ModelChangeable interface {
	Model
	Changeable
}

type Sortable interface {
	CanSort(key string) bool
}

type Pager interface {
	GetLimit() int
	GetPage() int
	GetSkip() int
	GetSort() string
	GetTotal() int
	SetTotal(n int)
	Sortable
}

// Sifter for select condition
type Sifter interface {
	Sift(q *SelectQuery) *SelectQuery
}

// SifterX with context
type SifterX interface {
	SiftX(ctx context.Context, q *SelectQuery) *SelectQuery
}

type ListArg interface {
	Pager
	Sifter
	Deleted() bool // select from trash schema like soft delete
}


```

Example
---

### define models

```go

import (
	"github.com/cupogo/andvari/models/comm"
	"github.com/cupogo/andvari/models/oid"
)

// Article 文章
type Article struct {
	comm.BaseModel `bun:"table:cms_article,alias:a" json:"-"`

	comm.DefaultModel

	ArticleBasic
} // @name Article

type ArticleBasic struct {
	// 作者
	Author string `bun:",notnull" extensions:"x-order=A" json:"author"`
	// 标题
	Title string `bun:",notnull" extensions:"x-order=B" json:"title"`
	// 内容
	Content string `bun:",notnull" extensions:"x-order=C" json:"content"`
} // @name ArticleBasic

type Articles []Article

// Creating function call to it's inner fields defined hooks
func (z *Article) Creating() error {
	if z.IsZeroID() {
		z.SetID(oid.NewID(oid.OtArticle))
	}

	return z.DefaultModel.Creating()
}

```

### database prepare testing
```sql
CREATE USER testing WITH LOGIN PASSWORD 'develop';
CREATE DATABASE testing WITH OWNER = testing ENCODING = 'UTF8';
GRANT ALL ON DATABASE testing TO testing;
```

### database store open

```go

RegisterModel((*cms1.Article)(nil))

dsn := "postgres://testing:develop0@localhost/testing?sslmode=disable"
tscfg := "zhcfg"
debug := false
db, err := pgx.Open(dsn ,tscfg, debug)

// create all registered tables
dropIt := false
err = db.InitSchemas(ctx, dropIt)
```

### data access

```go

// custom a Sifter
type ArticleSpec struct {
	pgx.PageSpec // Pager
	pgx.ModelSpec // Sifter

	// 作者
	Author string `extensions:"x-order=A" form:"author" json:"author"`
	// 标题
	Title string `extensions:"x-order=B" form:"title" json:"title"`
}

func (spec *ArticleSpec) Sift(q *ormQuery) *ormQuery {
	q = spec.ModelSpec.Sift(q)
	q, _ = siftICE(q, "author", spec.Author, false)
	q, _ = siftMatch(q, "title", spec.Title, false)

	return q
}
func (spec *ArticleSpec) CanSort(k string) bool {
	switch k {
	case "author":
		return true
	default:
		return spec.ModelSpec.CanSort(k)
	}
}


type contentStore struct {
	w *Wrap // a database wrapper
}

func (s *contentStore) ListArticle(ctx context.Context, spec *ArticleSpec) (data cms1.Articles, total int, err error) {
	total, err = s.w.db.ListModel(ctx, spec, &data)
	return
}
func (s *contentStore) GetArticle(ctx context.Context, id string) (obj *cms1.Article, err error) {
	obj = new(cms1.Article)
	err = s.w.db.GetModel(ctx, obj, id)
	return
}
func (s *contentStore) CreateArticle(ctx context.Context, in cms1.ArticleBasic) (obj *cms1.Article, err error) {
	obj = &cms1.Article{
		ArticleBasic: in,
	}
	err = s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
		if err = dbBeforeSaveArticle(ctx, tx, obj); err != nil {
			return err
		}
		dbOpModelMeta(ctx, tx, obj)
		err = dbInsert(ctx, tx, obj)
		return err
	})
	return
}
func (s *contentStore) UpdateArticle(ctx context.Context, id string, in cms1.ArticleSet) error {
	exist := new(cms1.Article)
	if err := dbGetWithPKID(ctx, s.w.db, exist, id); err != nil {
		return err
	}
	_ = exist.SetWith(in)
	return s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
		if err = dbBeforeSaveArticle(ctx, tx, exist); err != nil {
			return
		}
		dbOpModelMeta(ctx, tx, exist)
		return dbUpdate(ctx, tx, exist)
	})
}
func (s *contentStore) DeleteArticle(ctx context.Context, id string) error {
	obj := new(cms1.Article)
	if err := dbGetWithPKID(ctx, s.w.db, obj, id); err != nil {
		return err
	}
	return s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
		err = dbDeleteT(ctx, tx, s.w.db.Schema(), s.w.db.SchemaCrap(), "cms_article", obj.ID)
		if err != nil {
			return
		}
		return dbAfterDeleteArticle(ctx, tx, obj)
	})
}
```
