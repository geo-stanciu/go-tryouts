package main

import (
	"reflect"
	"strconv"
	"time"
)

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func InvokeMethodByName(any interface{}, name string, args ...interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}

	return reflect.ValueOf(any).MethodByName(name).Call(inputs)
}

func string2int(sval string) int {
	val, err := strconv.Atoi(sval)

	if err != nil {
		Log(true, err, "conversion", "Conversion error.", "val", val)
		return 0
	}

	return val
}
