package comm

import "github.com/cupogo/andvari/utils/array"

type ChangeMod struct {
	cs   array.String
	isUp bool
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
