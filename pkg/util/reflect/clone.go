package reflect

import "reflect"

// CloneNew 根据传入的对象克隆出一个全新对象
func CloneNew(i interface{}) interface{} {
	t := reflect.TypeOf(i)
	var tv reflect.Value
	switch t.Kind() {
	case reflect.Ptr:
		tv = reflect.New(t.Elem())
	default:
		tv = reflect.New(t)
	}
	ir := tv.Interface()
	return ir
}
