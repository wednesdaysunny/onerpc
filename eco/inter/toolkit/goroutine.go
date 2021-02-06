package toolkit

import (
	"fmt"
	"reflect"
)

type StdGoRoutine interface {
	Start()
}

func GO(goroutine StdGoRoutine) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				value := reflect.ValueOf(goroutine)
				fmt.Println("GO Error..xxxxxxxxxxx..FBI WARRING XXXXXX:", r, value)
			}
		}()
		goroutine.Start()
	}()
}
