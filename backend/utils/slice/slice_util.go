package sliceutil

import (
	. "cattail/backend/utils"
	"sort"
	"strconv"
	"strings"
)

func Get[S ~[]T, T any](arr S, index int, defaultVal T) T {
	if index < 0 || index >= len(arr) {
		return defaultVal
	}
	return arr[index]
}

func Remove[S ~[]T, T any](arr S, index int) S {
	return append(arr[:index], arr[index+1:]...)
}

func RemoveIf[S ~[]T, T any](arr S, cond func(T) bool) S {
	l := len(arr)
	if l <= 0 {
		return arr
	}
	for i := l - 1; i >= 0; i-- {
		if cond(arr[i]) {
			arr = append(arr[:i], arr[i+1:]...)
		}
	}
	return arr
}

func RemoveRange[S ~[]T, T any](arr S, from, to int) S {
	return append(arr[:from], arr[to:]...)
}

func Find[S ~[]T, T any](arr S, matchFunc func(int) bool) (int, bool) {
	total := len(arr)
	for i := 0; i < total; i++ {
		if matchFunc(i) {
			return i, true
		}
	}
	return -1, false
}

func AnyMatch[S ~[]T, T any](arr S, matchFunc func(int) bool) bool {
	total := len(arr)
	if total > 0 {
		for i := 0; i < total; i++ {
			if matchFunc(i) {
				return true
			}
		}
	}
	return false
}

func AllMatch[S ~[]T, T any](arr S, matchFunc func(int) bool) bool {
	total := len(arr)
	for i := 0; i < total; i++ {
		if !matchFunc(i) {
			return false
		}
	}
	return true
}

func Equals[S ~[]T, T comparable](arr1, arr2 S) bool {
	if &arr1 == &arr2 {
		return true
	}

	len1, len2 := len(arr1), len(arr2)
	if len1 != len2 {
		return false
	}
	for i := 0; i < len1; i++ {
		if arr1[i] != arr2[i] {
			return false
		}
	}
	return true
}

func Contains[S ~[]T, T Hashable](arr S, elem T) bool {
	return AnyMatch(arr, func(idx int) bool {
		return arr[idx] == elem
	})
}

func ContainsAny[S ~[]T, T Hashable](arr S, elems ...T) bool {
	for _, elem := range elems {
		if Contains(arr, elem) {
			return true
		}
	}
	return false
}

func ContainsAll[S ~[]T, T Hashable](arr S, elems ...T) bool {
	for _, elem := range elems {
		if !Contains(arr, elem) {
			return false
		}
	}
	return true
}

func Filter[S ~[]T, T any](arr S, filterFunc func(int) bool) []T {
	total := len(arr)
	var result []T
	for i := 0; i < total; i++ {
		if filterFunc(i) {
			result = append(result, arr[i])
		}
	}
	return result
}

func Map[S ~[]T, T any, R any](arr S, mappingFunc func(int) R) []R {
	total := len(arr)
	result := make([]R, total)
	for i := 0; i < total; i++ {
		result[i] = mappingFunc(i)
	}
	return result
}

func FilterMap[S ~[]T, T any, R any](arr S, mappingFunc func(int) (R, bool)) []R {
	total := len(arr)
	result := make([]R, 0, total)
	var filter bool
	var mapItem R
	for i := 0; i < total; i++ {
		if mapItem, filter = mappingFunc(i); filter {
			result = append(result, mapItem)
		}
	}
	return result
}

func ToMap[S ~[]T, T any, K Hashable, V any](arr S, mappingFunc func(int) (K, V)) map[K]V {
	total := len(arr)
	result := map[K]V{}
	for i := 0; i < total; i++ {
		key, val := mappingFunc(i)
		result[key] = val
	}
	return result
}

func Flat[T any](arr [][]T) []T {
	total := len(arr)
	var result []T
	for i := 0; i < total; i++ {
		subTotal := len(arr[i])
		for j := 0; j < subTotal; j++ {
			result = append(result, arr[i][j])
		}
	}
	return result
}

func FlatMap[T any, R any](arr [][]T, mappingFunc func(int, int) R) []R {
	total := len(arr)
	var result []R
	for i := 0; i < total; i++ {
		subTotal := len(arr[i])
		for j := 0; j < subTotal; j++ {
			result = append(result, mappingFunc(i, j))
		}
	}
	return result
}

func FlatValueMap[T Hashable](arr [][]T) []T {
	return FlatMap(arr, func(i, j int) T {
		return arr[i][j]
	})
}

func Reduce[S ~[]T, T any, R any](arr S, init R, reduceFunc func(R, T) R) R {
	result := init
	for _, item := range arr {
		result = reduceFunc(result, item)
	}
	return result
}

func Reverse[S ~[]T, T any](arr S) S {
	total := len(arr)
	for i := 0; i < total/2; i++ {
		arr[i], arr[total-i-1] = arr[total-i-1], arr[i]
	}
	return arr
}

func Join[S ~[]T, T any](arr S, sep string, toStringFunc func(int) string) string {
	total := len(arr)
	if total <= 0 {
		return ""
	}
	if total == 1 {
		return toStringFunc(0)
	}

	sb := strings.Builder{}
	for i := 0; i < total; i++ {
		if i != 0 {
			sb.WriteString(sep)
		}
		sb.WriteString(toStringFunc(i))
	}
	return sb.String()
}

func JoinString(arr []string, sep string) string {
	return Join(arr, sep, func(idx int) string {
		return arr[idx]
	})
}

func JoinInt(arr []int, sep string) string {
	return Join(arr, sep, func(idx int) string {
		return strconv.Itoa(arr[idx])
	})
}

func Unique[S ~[]T, T Hashable](arr S) S {
	result := make(S, 0, len(arr))
	uniKeys := map[T]struct{}{}
	var exists bool
	for _, item := range arr {
		if _, exists = uniKeys[item]; !exists {
			uniKeys[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func UniqueEx[S ~[]T, T any](arr S, toKeyFunc func(i int) string) S {
	result := make(S, 0, len(arr))
	keyArr := Map(arr, toKeyFunc)
	uniKeys := map[string]struct{}{}
	var exists bool
	for i, item := range arr {
		if _, exists = uniKeys[keyArr[i]]; !exists {
			uniKeys[keyArr[i]] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func Sort[S ~[]T, T Hashable](arr S) S {
	sort.Slice(arr, func(i, j int) bool {
		return arr[i] <= arr[j]
	})
	return arr
}

func SortDesc[S ~[]T, T Hashable](arr S) S {
	sort.Slice(arr, func(i, j int) bool {
		return arr[i] > arr[j]
	})
	return arr
}

func Union[S ~[]T, T Hashable](arr1 S, arr2 S) S {
	hashArr, compArr := arr1, arr2
	if len(arr1) < len(arr2) {
		hashArr, compArr = compArr, hashArr
	}
	hash := map[T]struct{}{}
	for _, item := range hashArr {
		hash[item] = struct{}{}
	}

	uniq := map[T]struct{}{}
	ret := make(S, 0, len(compArr))
	exists := false
	for _, item := range compArr {
		if _, exists = hash[item]; exists {
			if _, exists = uniq[item]; !exists {
				ret = append(ret, item)
				uniq[item] = struct{}{}
			}
		}
	}
	return ret
}

func Exclude[S ~[]T, T Hashable](arr1 S, arr2 S) S {
	diff := make([]T, 0, len(arr1))
	hash := map[T]struct{}{}
	for _, item := range arr2 {
		hash[item] = struct{}{}
	}

	for _, item := range arr1 {
		if _, exists := hash[item]; !exists {
			diff = append(diff, item)
		}
	}
	return diff
}

func PadLeft[S ~[]T, T any](arr S, val T, count int) S {
	prefix := make(S, count)
	for i := 0; i < count; i++ {
		prefix[i] = val
	}
	arr = append(prefix, arr...)
	return arr
}

func PadRight[S ~[]T, T any](arr S, val T, count int) S {
	for i := 0; i < count; i++ {
		arr = append(arr, val)
	}
	return arr
}

func RemoveLeft[S ~[]T, T comparable](arr S, val T) S {
	for len(arr) > 0 && arr[0] == val {
		arr = arr[1:]
	}
	return arr
}

func RemoveRight[S ~[]T, T comparable](arr S, val T) S {
	for {
		length := len(arr)
		if length > 0 && arr[length-1] == val {
			arr = arr[:length]
		} else {
			break
		}
	}
	return arr
}

func Count[S ~[]T, T any](arr S, filter func(int) bool) int {
	count := 0
	for i := range arr {
		if filter(i) {
			count += 1
		}
	}
	return count
}

func Group[S ~[]T, T any, K Hashable, R any](arr S, groupFunc func(int) (K, R)) map[K][]R {
	ret := map[K][]R{}
	for i := range arr {
		key, val := groupFunc(i)
		ret[key] = append(ret[key], val)
	}
	return ret
}
