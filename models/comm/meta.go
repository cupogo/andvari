package comm

// JsonKV as meta values
type JsonKV map[string]any // @name JsonKV
// Meta alias of map
type Meta = JsonKV // @name Meta

// IsEmpty ...
func (m JsonKV) IsEmpty() bool {
	return len(m) == 0
}

// Merge ...
func (m *JsonKV) Merge(other JsonKV) {
	*m = MergeMeta(*m, other)
}

// MergeMeta merge a map to other
func MergeMeta(m, o JsonKV) JsonKV {
	if m == nil {
		m = JsonKV{}
	}
	if o == nil {
		return m
	}
	for k, v := range o {
		m[k] = v
	}
	return m
}

// Get ...
func (m JsonKV) Get(key string) (v any, ok bool) {
	v, ok = m[key]
	return
}

func (m JsonKV) GetInt(key string) int {
	if v, ok := m[key]; ok {
		if s, ok := v.(int); ok {
			return s
		}
	}
	return 0
}

func (m JsonKV) GetStr(key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Set ...
func (m JsonKV) Set(k string, v any) {
	if m == nil {
		m = JsonKV{}
	}
	m[k] = v
}

// Unset ...
func (m JsonKV) Unset(k string) {
	delete(m, k)
}

// Filter ...
func (m JsonKV) Filter(keys ...string) (out JsonKV) {
	out = JsonKV{}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			out[k] = v
		}
	}
	return
}

// Copy ...
func (m JsonKV) Copy() (out JsonKV) {
	out = JsonKV{}
	for k, v := range m {
		out[k] = v
	}
	return
}

type KV struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type KVs []KV

type MetaDiff struct {
	Add    KVs      `json:"add"` // 批量添加/更新
	Delete []string `json:"del"` // 批量删除
}

type MetaField struct {
	// Meta 元信息
	Meta JsonKV `bson:"meta,omitempty" json:"meta,omitempty" pg:"meta,notnull,use_zero,default:'{}'" swaggerignore:"true"`
}

func (mf *MetaField) MergeMeta(other JsonKV) {
	mf.Meta = MergeMeta(mf.Meta, other)
}

func (mf *MetaField) MetaGet(key string) (v any, ok bool) {
	if mf.Meta != nil {
		return mf.Meta.Get(key)
	}
	return
}

func (mf *MetaField) MetaSet(key string, value any) {
	if mf.Meta == nil {
		mf.Meta = JsonKV{}
	}
	mf.Meta[key] = value
}

func (mf *MetaField) MetaUnset(key string) {
	if mf.Meta != nil {
		mf.Meta.Unset(key)
	}
}

func (mf *MetaField) UpMeta(up *MetaDiff) (ok bool) {
	if up == nil {
		return false
	}

	for _, k := range up.Delete {
		if len(k) > 0 {
			mf.MetaUnset(k)
			ok = true
		}
	}

	for _, i := range up.Add {
		if len(i.Key) > 0 && i.Value != nil {
			mf.MetaSet(i.Key, i.Value)
			ok = true
		}
	}

	return
}

func (mu *MetaDiff) AddKV(k string, v any) {
	if mu == nil {
		mu = &MetaDiff{}
	}
	if mu.Add == nil {
		mu.Add = KVs{}
	}
	mu.Add = append(mu.Add, KV{
		Key: k, Value: v,
	})
}
