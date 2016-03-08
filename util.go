package main

import "reflect"

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func CmpU8Arr(a, b []uint8) bool {
	if len(a) == len(b) {
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	return false
}

func Mux(out interface{}, in ...interface{}) {
	var inarr []reflect.SelectCase
	if len(in) == 1 && reflect.TypeOf(in[0]).Kind() == reflect.Slice {
		nin := reflect.ValueOf(in[0])
		inarr = make([]reflect.SelectCase, nin.Len())
		for i := range inarr {
			inarr[i] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: nin.Index(i),
			}
		}
	} else {
		inarr = make([]reflect.SelectCase, len(in))
		for i := range in {
			inarr[i] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(in[i]),
			}
		}
	}
	ovalue := reflect.ValueOf(out)
	done := len(inarr)
	for done > 0 {
		i, v, ok := reflect.Select(inarr)
		if !ok {
			inarr[i].Chan = reflect.ValueOf(nil)
			done--
			continue
		}
		ovalue.Send(v)
	}
	ovalue.Close()
}
