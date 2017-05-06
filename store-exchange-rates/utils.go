package main

import "strconv"

func string2int(sval string) int {
	val, err := strconv.Atoi(sval)

	if err != nil {
		return 0
	}

	return val
}
