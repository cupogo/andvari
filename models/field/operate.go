package field

import "fmt"

// 操作类型
type OperateType int8

const (
	OperateTypeUpdate OperateType = 1 << iota //  1 修改
	OperateTypeDelete                         //  2 删除
	OperateTypeCreate                         //  4 新增
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
