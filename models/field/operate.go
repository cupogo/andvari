package field

import "fmt"

// 操作类型
type OperateType int8

type ModelOperateType = OperateType // deprecated

const (
	OperateTypeCreate OperateType = 1 << iota //  1 新增
	OperateTypeUpdate                         //  2 修改
	OperateTypeDelete                         //  4 删除
	OperateTypeCustom                         //  8 手动或其他
)

func (z *OperateType) Decode(s string) error {
	switch s {
	case "1", "create":
		*z = OperateTypeCreate
	case "2", "update":
		*z = OperateTypeUpdate
	case "4", "delete":
		*z = OperateTypeDelete
	case "8", "custom", "other":
		*z = OperateTypeCustom
	default:
		return fmt.Errorf("invalid OperateType: %q", s)
	}
	return nil
}
func (z OperateType) String() string {
	switch z {
	case OperateTypeCreate:
		return "create"
	case OperateTypeUpdate:
		return "update"
	case OperateTypeDelete:
		return "delete"
	case OperateTypeCustom:
		return "custom"
	default:
		return fmt.Sprintf("OperateType%02x", int8(z))
	}
}
func (z OperateType) MarshalText() ([]byte, error) {
	return []byte(z.String()), nil
}
