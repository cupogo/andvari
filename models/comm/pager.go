package comm

type Pager interface {
	GetLimit() int
	GetPage() int
	GetSkip() int
	GetSort() string
	GetTotal() int
	SetTotal(n int)
}

// PageSpec ...
type PageSpec struct {
	// 分页大小，默认20
	Limit int `json:"limit,omitempty" form:"limit" extensions:"x-order=x"`
	// 第几页
	Page int `json:"page,omitempty" form:"page" extensions:"x-order=y"`
	// 跳过多少条记录，如果提供 page 此项跳过
	Skip int `json:"skip,omitempty" form:"skip" extensions:"x-order=z"`
	// 排序 field desc | [asc]
	Sort string `json:"sort,omitempty" form:"sort" extensions:"x-order=|"`

	Total int `json:"total,omitempty" swaggerignore:"true"`
}

func (p *PageSpec) GetLimit() int {
	return p.Limit
}

func (p *PageSpec) GetPage() int {
	if p.Page == 0 && p.Limit > 0 {
		return p.Skip / p.Limit
	}
	return p.Page
}

// TODO pagespec validation

func (p *PageSpec) GetSkip() int {
	return p.Skip
}

func (p *PageSpec) SetSkip(skip int) {
	p.Skip = skip
}

// GetSort eg. createtime asc => {"createtime":-1}
func (p *PageSpec) GetSort() string {
	return p.Sort
}

func (p *PageSpec) GetTotal() int {
	return p.Total
}

func (p *PageSpec) SetTotal(n int) {
	p.Total = n
}
