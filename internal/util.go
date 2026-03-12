package internal

import "reflect"

//
// 存放公共工具函数
//

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}

	// 获取接口的反射值
	v := reflect.ValueOf(i)

	// 只有特定类型（指针、切片、映射、通道、函数、接口）才能判断 IsNil
	// 如果是 int, string 等基础类型，调用 IsNil 会 panic，所以要先判断 Kind
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}

	// 如果是值类型（如 int, struct），它不可能为 nil
	return false
}
