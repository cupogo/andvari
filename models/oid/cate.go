package oid

// ObjType 目标类型
type ObjType int16

// consts
const (
	OtDefault    ObjType = iota
	OtAccount            // 账号
	OtCompany            // 公司、企业
	OtDepartment         // 部门
	OtArticle            // 内容、文章、条款
	OtTeam               // 小组、群
	OtEvent              // 事件：任务、消息、日志等
	OtToken              // 票据
	OtPeople             // 人员: 客户信息、联系人、地址等
	OtForm               // 表单: 配置、订单、票据等
	OtGoods              // 东西: 设备、配件、软件等
	otLast
)

func (ot ObjType) Code() string {
	switch ot {
	case OtAccount:
		return "ac"
	case OtCompany:
		return "co"
	case OtDepartment:
		return "dp"
	case OtArticle:
		return "at"
	case OtTeam:
		return "tm"
	case OtToken:
		return "tk"
	case OtEvent:
		return "ev"
	case OtForm:
		return "fm"
	case OtPeople:
		return "pe"
	case OtGoods:
		return "go"
	}
	return "oo" // default
}
