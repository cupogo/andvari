package field

import "fmt"

// 操作类型
type OperateType int8

type ModelOperateType = OperateType // deprecated

const (
	OperateTypeCreate OperateType = 1 << iota //  1 新增
	OperateTypeUpdate                         //  2 修改
	OperateTypeDelete                         //  4 删除
)

func (z *OperateType) Decode(s string) error {
	switch s {
	case "1", "update":
		*z = OperateTypeUpdate
	case "2", "delete":
		*z = OperateTypeDelete
	case "4", "create":
		*z = OperateTypeCreate
	default:
		return fmt.Errorf("invalid OperateType: %q", s)
	}
	return nil
}
func (z OperateType) String() string {
	switch z {
	case OperateTypeUpdate:
		return "update"
	case OperateTypeDelete:
		return "delete"
	case OperateTypeCreate:
		return "create"
	default:
		return fmt.Sprintf("OperateType %d", int8(z))
	}
}
func (z OperateType) MarshalText() ([]byte, error) {
	return []byte(z.String()), nil
}
