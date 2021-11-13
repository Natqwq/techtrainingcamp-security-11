package main

import (
	"encoding/base64"
	"reflect"
	"unsafe"
)

func encode(data string) string {
	content := *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&data))))
	coder := base64.NewEncoding(Base64Table)
	return coder.EncodeToString(content)
}

func decode(data string) string {
	coder := base64.NewEncoding(Base64Table)
	result, _ := coder.DecodeString(data)
	return *(*string)(unsafe.Pointer(&result))
}
