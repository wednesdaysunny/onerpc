package reflect

import (
	"reflect"
	"runtime"
	"strings"
)

// MethodName fetches the exact name of a method
func MethodName(fn interface{}) string {
	method := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	parts := strings.Split(method, ".")
	if len(parts) < 1 {
		return ""
	}
	return parts[len(parts)-1]
}

// FullName fetches the full name of a method including the package and struct
func FullName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

// PbMethodName returns the method name for a given proto buffer method
func PbMethodName(fn interface{}) string {
	fullname := FullName(fn)
	parts := strings.Split(fullname, ".")
	if len(parts) < 2 {
		return ""
	}
	parts = parts[len(parts)-2:]
	tryTrim := func(name string, pf string) string {
		lname, lpf := len(name), len(pf)
		if lname > lpf {
			if name[lname-lpf:] == pf {
				return name[:lname-lpf]
			}
		}
		return ""
	}

	if svcname := tryTrim(parts[0], "Server"); svcname != "" {
		parts[0] = svcname
		return strings.Join(parts, ".")
	} else if svcname := tryTrim(parts[0], "Client"); svcname != "" {
		parts[0] = svcname
		return strings.Join(parts, ".")
	}
	return ""
}

// Contains checks whether an element is in a collection like Slice, Array, Map.
func Contains(target interface{}, obj interface{}) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}

	return false
}

func structValueCopy(src interface{}, desc interface{}) {
	srcRef := reflect.ValueOf(src).Elem()
	descRef := reflect.ValueOf(desc).Elem()
	srcTypeOf := srcRef.Type()
	for i := 0; i < srcRef.NumField(); i++ {
		field := srcRef.Field(i)
		fieldName := srcTypeOf.Field(i).Name

		if descRef.FieldByName(fieldName).Kind() != reflect.Invalid && descRef.FieldByName(fieldName).Type() == field.Type() {
			descRef.FieldByName(fieldName).Set(reflect.ValueOf(field.Interface()))
		}
	}
}
