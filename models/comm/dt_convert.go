package comm

// NewDateFromDateTime 将 DateTime 转换为 Date 类型
func NewDateFromDateTime(dt DateTime) Date {
	// 使用 DateTime 的 Time() 方法获取时间对象
	t := dt.Time()
	// 使用现有的 NewDateFromTime 函数转换为 Date
	return NewDateFromTime(t)
}

// NewDateTimeFromDate 将 Date 转换为 DateTime 类型
func NewDateTimeFromDate(d Date) DateTime {
	// 使用 Date 的 Time() 方法获取时间对象
	t := d.Time()
	// 使用现有的 NewDateTimeFromTime 函数转换为 DateTime
	return NewDateTimeFromTime(t)
}

func (dt DateTime) ToDate() Date {
	return NewDateFromDateTime(dt)
}

func (d Date) ToDateTime() DateTime {
	return NewDateTimeFromDate(d)
}
