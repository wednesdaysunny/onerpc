package intutil

import (
	"strconv"
)

func ToInt32(str string, defaultValue int32) int32 {
	if b, err := strconv.ParseInt(str, 10, 32); err != nil {
		return defaultValue
	} else {
		return int32(b)
	}
}

func ToInt64(str string, defaultValue int64) int64 {
	if b, err := strconv.ParseInt(str, 10, 64); err != nil {
		return defaultValue
	} else {
		return int64(b)
	}
}

func ToUint64(str string, defaultValue uint64) uint64 {
	if b, err := strconv.Atoi(str); err != nil {
		return defaultValue
	} else {
		return uint64(b)
	}
}

func DiffInt64(arr1 []int64, arr2 []int64) []int64 {
	res := make([]int64, 0)
	for _, val1 := range arr1 {
		exist := false
		for _, val2 := range arr2 {
			if val1 == val2 {
				exist = true
				break
			}
		}
		if !exist {
			res = append(res, val1)
		}
	}
	return res
}

func IntersectInt64(arr1 []int64, arr2 []int64) []int64 {
	res := make([]int64, 0)
	for _, val1 := range arr1 {
		for _, val2 := range arr2 {
			if val1 == val2 {
				res = append(res, val1)
				break
			}
		}
	}
	return res
}

func MaxUint64(x uint64, y uint64) uint64 {
	if x < y {
		return y
	}
	return x
}

func MinUint64(x uint64, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}

func MaxInt64(x int64, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

func MinInt64(x int64, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func FloatToPriceInt64(x float64) int64 {
	return int64(x * 100)
}

func IndexInt64(arr []int64, key int64) int {
	for index, val := range arr {
		if val == key {
			return index
		}
	}
	return -1
}
