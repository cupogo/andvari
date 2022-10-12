package comm

import (
	"strconv"
	"time"

	"github.com/cupogo/andvari/models/oid"
)

type OID = oid.OID

// IDField struct contain model's ID field.
type IDField struct {
	ID OID `bson:"_id,omitempty" json:"id" form:"id" bun:",pk,type:bigint" pg:",pk,type:bigint" extensions:"x-order=/"` // 主键
}

// DateFields struct contain `createdAt` and `updatedAt`
// fields that autofill on insert/update model.
type DateFields struct {
	CreatedAt time.Time `bson:"createdAt" json:"createdAt" form:"createdAt" bun:"created,notnull,default:now()" pg:"created,notnull,default:now()" extensions:"x-order=["` // 创建时间
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt" form:"updatedAt" bun:"updated,notnull" pg:"updated,notnull" extensions:"x-order=]"`                             // 变更时间
}

// PrepareID method prepare id value to using it as id in filtering,...
// e.g convert hex-string id value to bson.ObjectId
func (f *IDField) PrepareID(id any) (any, error) {
	if v := oid.Cast(id); v.Valid() {
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

func (f *DateFields) SetCreated(ts any) bool {
	if v, ok := AsTime(ts); ok {
		f.CreatedAt = v
		return ok
	}

	return false
}

// GetUpdated return time of updatedAt
func (f *DateFields) GetUpdated() time.Time {
	return f.UpdatedAt
}

type CreatorField struct {
	// 创建者ID
	CreatorID OID `bson:"creatorID,omitempty" json:"creatorID,omitempty" form:"creatorID" bun:"creator_id,notnull,default:0" pg:"creator_id,notnull,use_zero" extensions:"x-order=_"`
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

type OwnerField struct {
	// 所有者OID 默认为当前登录账号主键
	OwnerID OID `bson:"ownerID,omitempty" json:"ownerID,omitempty" form:"ownerID" bun:"owner_id,notnull,default:0" pg:"owner_id,notnull,use_zero" extensions:"x-order=@"`
}

// GetOwnerID 返回所有者ID
func (f *OwnerField) GetOwnerID() OID {
	return f.OwnerID
}

// SetOwnerID 设置所有者ID
func (f *OwnerField) SetOwnerID(id any) bool {
	if v := oid.Cast(id); v.Valid() {
		f.OwnerID = v
		return true
	}
	return false
}

// DefaultModel struct contain model's default fields.
type DefaultModel struct {
	IDField      `bson:",inline"`
	DateFields   `bson:",inline"`
	CreatorField `bson:",inline"`
	ChangeMod    `bson:",inline"`
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
	ID string `bson:"_id,omitempty" json:"id" form:"id" bun:",pk" pg:",pk" extensions:"x-order=/"` // 主键
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
	ChangeMod    `bson:",inline"`
}

// SerialField struct contain model's ID field.
type SerialField struct {
	ID int `bson:"_id,omitempty" json:"id" form:"id" bun:",pk,autoincrement" pg:",pk,type:serial" extensions:"x-order=/"` // 主键
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
	ChangeMod    `bson:",inline"`
}

// Creating function call to it's inner fields defined hooks
func (model *SerialModel) Creating() error {
	return model.DateFields.Creating()
}

// Saving function call to it's inner fields defined hooks
func (model *SerialModel) Saving() error {
	return model.DateFields.Saving()
}

type TextSearchField struct {
	// 生成 tsvector 时所使用的配置名
	TsCfgName string `json:"-" bun:"ts_cfg,notnull,type:name" pg:"ts_cfg,notnull,default:'',type:name"`
	// tsvector 格式的关键词，用于全文检索
	TsVector string `bson:"textKeyword" json:"-" bun:"ts_vec,type:tsvector" pg:"ts_vec,type:tsvector"`

	cols []string
} // @name TextSearchField

func (tsf *TextSearchField) GetTsConfig() string {
	return tsf.TsCfgName
}

func (tsf *TextSearchField) SetTsColumns(cols ...string) {
	tsf.cols = cols
}

func (tsf *TextSearchField) GetTsColumns() []string {
	return tsf.cols
}
