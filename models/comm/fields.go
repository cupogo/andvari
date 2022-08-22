package comm

import (
	"strconv"
	"time"

	"hyyl.xyz/cupola/andvari/models/idgen"
	"hyyl.xyz/cupola/andvari/models/oid"
)

type IID = idgen.IID
type OID = oid.OID
type IIDs = []IID
type OIDs = []OID

// IDField struct contain model's ID field.
type IDField struct {
	ID OID `bson:"_id,omitempty" json:"id" form:"id" pg:",pk,type:bigint" extensions:"x-order=/"` // 主键
}

// DateFields struct contain `createdAt` and `updatedAt`
// fields that autofill on insert/update model.
type DateFields struct {
	CreatedAt time.Time `bson:"createdAt" json:"createdAt" form:"createdAt" pg:"created,notnull,default:now()" extensions:"x-order=k"` // 创建时间
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt" form:"updatedAt" pg:"updated,notnull" extensions:"x-order=l"`               // 变更时间
}

// PrepareID method prepare id value to using it as id in filtering,...
// e.g convert hex-string id value to bson.ObjectId
func (f *IDField) PrepareID(id any) (any, error) {
	if v, ok := id.(OID); ok {
		return v, nil
	}
	if v, ok := id.(IID); ok {
		return v, nil
	}

	// Otherwise id must be ObjectId
	return id, nil
}

// GetID method return model's id
func (f *IDField) GetID() any {
	return f.ID
}

// SetID set id value of model's id field.
func (f *IDField) SetID(id any) bool {
	if v := oid.Cast(id); v.Valid() {
		f.ID = v
		return true
	}
	return false
}

// IsZeroID check id is empty
func (f *IDField) IsZeroID() bool {
	return f.ID.IsZero()
}

func (f *IDField) StringID() string {
	return f.ID.String()
}

//--------------------------------
// DateField methods
//--------------------------------

// Creating hook used here to set `created` field
// value on inserting new model into database.
func (f *DateFields) Creating() error {
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now()
	}

	return nil
}

// Saving hook used here to set `updated` field value
// on create/update model.
func (f *DateFields) Saving() error {
	f.UpdatedAt = time.Now()

	return nil
}

// GetUpdated return time of updatedAt
func (f *DateFields) GetUpdated() time.Time {
	return f.UpdatedAt
}

type CreatorField struct {
	// 创建者ID
	CreatorID OID `bson:"creatorID,omitempty" json:"creatorID,omitempty" form:"creatorID" pg:"creator_id,notnull,use_zero" extensions:"x-order=m"`
}

// GetCreatorID 返回创建者ID
func (f *CreatorField) GetCreatorID() OID {
	return f.CreatorID
}

// SetCreatorID 设置创建者ID
func (f *CreatorField) SetCreatorID(id any) bool {
	if v := oid.Cast(id); v.Valid() {
		f.CreatorID = v
		return true
	}
	return false
}

// DefaultModel struct contain model's default fields.
type DefaultModel struct {
	IDField      `bson:",inline"`
	DateFields   `bson:",inline"`
	CreatorField `bson:",inline"`
	ChangeMod
}

// Creating function call to it's inner fields defined hooks
func (model *DefaultModel) Creating() error {
	return model.DateFields.Creating()
}

// Saving function call to it's inner fields defined hooks
func (model *DefaultModel) Saving() error {
	return model.DateFields.Saving()
}

type IDFieldStr struct {
	ID string `bson:"_id,omitempty" json:"id" form:"id" pg:",pk" extensions:"x-order=/"` // 主键
}

func (f *IDFieldStr) PrepareID(id any) (any, error) {
	if v, ok := id.(string); ok {
		return v, nil
	}

	// Otherwise id must be ObjectId
	return id, nil
}

// GetID method return model's id
func (f *IDFieldStr) GetID() any {
	return f.ID
}

// SetID set id value of model's id field.
func (f *IDFieldStr) SetID(id any) bool {
	if v, ok := id.(string); ok {
		f.ID = v
		return len(v) > 0
	}
	return false
}

// IsZeroID check id is empty
func (f *IDFieldStr) IsZeroID() bool {
	return len(f.ID) == 0
}

func (f *IDFieldStr) StringID() string {
	return f.ID
}

// DunceModel struct contain model's default fields with string pk.
type DunceModel struct {
	IDFieldStr   `bson:",inline"`
	DateFields   `bson:",inline"`
	CreatorField `bson:",inline"`
	ChangeMod
}

// SerialField struct contain model's ID field.
type SerialField struct {
	ID int `bson:"_id,omitempty" json:"id" form:"id" pg:",pk,type:serial" extensions:"x-order=/"` // 主键
}

func (f *SerialField) PrepareID(id any) (any, error) {
	if v, ok := id.(int); ok {
		return v, nil
	}

	// Otherwise id must be ObjectId
	return id, nil
}

// GetID method return model's id
func (f *SerialField) GetID() any {
	return f.ID
}

// SetID set id value of model's id field.
func (f *SerialField) SetID(id any) bool {
	if v, ok := id.(int); ok {
		f.ID = v
		return v > 0
	}
	if v, ok := id.(string); ok {
		var err error
		f.ID, _ = strconv.Atoi(v)
		return err == nil && f.ID > 0
	}
	return false
}

// IsZeroID check id is empty
func (f *SerialField) IsZeroID() bool {
	return f.ID == 0
}

func (f *SerialField) StringID() string {
	return strconv.Itoa(f.ID)
}

// SerialModel struct contain model's default fields.
type SerialModel struct {
	SerialField  `bson:",inline"`
	DateFields   `bson:",inline"`
	CreatorField `bson:",inline"`
	ChangeMod
}

// Creating function call to it's inner fields defined hooks
func (model *SerialModel) Creating() error {
	return model.DateFields.Creating()
}

// Saving function call to it's inner fields defined hooks
func (model *SerialModel) Saving() error {
	return model.DateFields.Saving()
}

// StringsDiff deprecated
type StringsDiff struct {
	Newest  []string `json:"newest" validate:"dive"`  // 新增的字串集
	Removed []string `json:"removed" validate:"dive"` // 删除的字串集
} // @name StringsDiff

type TextSearchField struct {
	// 生成 tsvector 时所使用的配置名
	TsCfgName string `json:"-" pg:"ts_cfg,notnull,use_zero,type:name"`
	// tsvector 格式的关键词，用于全文检索
	TsVector string `bson:"textKeyword" json:"-" pg:"ts_vec,type:tsvector"`
} // @name TextSearchField

// GenSlug deprecated
func GenSlug() string {
	return oid.NewID(oid.OtDefault).String()
}
