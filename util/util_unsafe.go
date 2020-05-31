// +build !appengine,!js

package util

import (
	"reflect"
	"runtime"
	"unsafe"
)

// BytesToReadOnlyString returns a string converted from given bytes.
func BytesToReadOnlyString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToReadOnlyBytes returns bytes converted from given string.
func StringToReadOnlyBytes(s string) []byte {
	b := make([]byte, 0)
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	runtime.KeepAlive(s)
	return b
}
