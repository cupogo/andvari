package comm

import "fmt"

// ExampleTime 提供了一些测试用例
func ExampleTime() {
	t1, _ := ParseTime("12:34:56.789")
	fmt.Println(t1)                    // 输出: 12:34:56.789
	fmt.Println(t1.Format("hh:mm:ss")) // 输出: 12:34:56

	t2, _ := ParseTime("00:00:00.000")
	fmt.Println(t2.IsZero()) // 输出: true

	t3, _ := ParseTime("12:34")
	fmt.Println(t3) // 输出: 12:34

	bytes, _ := t1.MarshalText()
	fmt.Println(string(bytes)) // 输出: 12:34:56.789

	var t4 Time
	_ = t4.UnmarshalText(bytes)
	fmt.Println(t4) // 输出: 12:34:56.789

	// Output:
	// 12:34:56.789
	// 12:34:56
	// true
	// 12:34
	// 12:34:56.789
	// 12:34:56.789
}
