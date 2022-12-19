package field

import "fmt"

// ORM操作类型
type ModelOperateType int8

const (
	ModelOperateTypeUpdate ModelOperateType = 1 << iota //  1 修改
	ModelOperateTypeDelete                              //  2 删除
	ModelOperateTypeCreate                              //  4 新增
)

func (z *ModelOperateType) Decode(s string) error {
	switch s {
	case "1", "update":
		*z = ModelOperateTypeUpdate
	case "2", "delete":
		*z = ModelOperateTypeDelete
	case "4", "create":
		*z = ModelOperateTypeCreate
	default:
		return fmt.Errorf("invalid ModelOperateType: %q", s)
	}
	return nil
}
func (z ModelOperateType) String() string {
	switch z {
	case ModelOperateTypeUpdate:
		return "update"
	case ModelOperateTypeDelete:
		return "delete"
	case ModelOperateTypeCreate:
		return "create"
	default:
		return fmt.Sprintf("ModelOperateType %d", int8(z))
	}
}
