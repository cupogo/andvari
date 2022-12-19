package comm

import (
	"github.com/cupogo/andvari/utils/array"
)

type ChangeValue struct {
	// 列名称
	Column string `bson:"column,omitempty" json:"column"  `
	// 旧值
	OldVal any `bson:"oldVal,omitempty" json:"oldVal"  `
	// 新值
	NewVal any `bson:"newVal,omitempty" json:"newVal" `
}

type ChangeValues []ChangeValue

type ChangeMod struct {
	cs   array.String
	isUp bool
	cv   ChangeValues
}

func (cm *ChangeMod) SetChange(cs ...string) {
	if cm.cs == nil {
		cm.cs = array.NewString(cs...)
	} else {
		cm.cs.Insert(cs...)
	}
}

func (cm *ChangeMod) GetChanges() []string {
	return cm.cs.UnsortedList()
}

func (cm *ChangeMod) CountChange() int {
	return len(cm.cs)
}

func (cm *ChangeMod) HasChange(name string) bool {
	return cm.cs.Has(name)
}

func (cm *ChangeMod) IsUpdate() bool {
	return cm.isUp
}

func (cm *ChangeMod) SetIsUpdate(v bool) {
	cm.isUp = v
}

func (cm *ChangeMod) LogChangeValue(c string, ov, nv any) {
	if !cm.HasChange(c) {
		cm.cv = append(cm.cv, ChangeValue{Column: c, OldVal: ov, NewVal: nv})
	}
	cm.SetChange(c)
}

func (cm *ChangeMod) ChangedValues() ChangeValues {
	return cm.cv
}

func (cm *ChangeMod) IsLog() bool {
	return true
}
