package utils

import "reflect"

// /////////////////////////////////////////////////
// 插入元素
func SliceInsertI32(slice []uint32, val uint32, idx int) []uint32 {
	rear := append([]uint32{}, slice[idx:]...)
	slice = append(slice[0:idx], val)
	slice = append(slice, rear...)
	return slice
}
func SliceInsertI64(slice []int64, val int64, idx int) []int64 {
	rear := append([]int64{}, slice[idx:]...)
	slice = append(slice[0:idx], val)
	slice = append(slice, rear...)
	return slice
}
func SliceInsertStr(slice []string, val string, idx int) []string {
	rear := append([]string{}, slice[idx:]...)
	slice = append(slice[0:idx], val)
	slice = append(slice, rear...)
	return slice
}

// /////////////////////////////////////////////////
// 删除元素
func SliceDelI32(slice []uint32, val uint32) []uint32 {
	for k, v := range slice {
		if v != val {
			continue
		}
		slice = append(slice[:k], slice[k+1:]...)
	}
	return slice
}
func SliceDelI64(slice []int64, val int64) []int64 {
	for k, v := range slice {
		if v != val {
			continue
		}
		slice = append(slice[:k], slice[k+1:]...)
	}
	return slice
}
func SliceDelStr(slice []string, val string) []string {
	for k, v := range slice {
		if v != val {
			continue
		}
		slice = append(slice[:k], slice[k+1:]...)
	}
	return slice
}

// /////////////////////////////////////////////////
// 求交集
func SliceInterI32(slice1, slice2 []uint32) []uint32 {
	m := make(map[uint32]int)
	nn := make([]uint32, 0)
	for _, v := range slice1 {
		m[v]++
	}
	for _, v := range slice2 {
		times, _ := m[v]
		if times > 0 {
			nn = append(nn, v)
		}
	}
	return nn
}
func SliceInterI64(slice1, slice2 []int64) []int64 {
	m := make(map[int64]int)
	nn := make([]int64, 0)
	for _, v := range slice1 {
		m[v]++
	}
	for _, v := range slice2 {
		times, _ := m[v]
		if times > 0 {
			nn = append(nn, v)
		}
	}
	return nn
}
func SliceInterStr(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	for _, v := range slice1 {
		m[v]++
	}
	for _, v := range slice2 {
		times, _ := m[v]
		if times > 0 {
			nn = append(nn, v)
		}
	}
	return nn
}

// /////////////////////////////////////////////////
// 求并集
func SliceUnionI32(slice1, slice2 []uint32) []uint32 {
	m := make(map[uint32]int)
	for _, v := range slice1 {
		m[v]++
	}
	for _, v := range slice2 {
		times, _ := m[v]
		if times == 0 {
			slice1 = append(slice1, v)
		}
	}
	return slice1
}
func SliceUnionI64(slice1, slice2 []int64) []int64 {
	m := make(map[int64]int)
	for _, v := range slice1 {
		m[v]++
	}
	for _, v := range slice2 {
		times, _ := m[v]
		if times == 0 {
			slice1 = append(slice1, v)
		}
	}
	return slice1
}
func SliceUnionIStr(slice1, slice2 []string) []string {
	m := make(map[string]int)
	for _, v := range slice1 {
		m[v]++
	}
	for _, v := range slice2 {
		times, _ := m[v]
		if times == 0 {
			slice1 = append(slice1, v)
		}
	}
	return slice1
}

// /////////////////////////////////////////////////
// 求补集
func SliceCompleI32(slice1, slice2 []uint32) []uint32 {
	m := make(map[uint32]int)
	for _, v := range slice1 {
		m[v]++
	}
	for _, v := range slice2 {
		m[v]++
	}
	nn := make([]uint32, 0)
	for value, num := range m {
		if num == 1 {
			nn = append(nn, value)
		}
	}
	return nn
}
func SliceCompleI64(slice1, slice2 []int64) []int64 {
	m := make(map[int64]int)
	for _, v := range slice1 {
		m[v]++
	}
	for _, v := range slice2 {
		m[v]++
	}
	nn := make([]int64, 0)
	for value, num := range m {
		if num == 1 {
			nn = append(nn, value)
		}
	}
	return nn
}
func SliceCompleStr(slice1, slice2 []string) []string {
	m := make(map[string]int)
	for _, v := range slice1 {
		m[v]++
	}
	for _, v := range slice2 {
		m[v]++
	}
	nn := make([]string, 0)
	for value, num := range m {
		if num == 1 {
			nn = append(nn, value)
		}
	}
	return nn
}

// /////////////////////////////////////////////////
// 求差集（slice1-交集）
func SliceDiffI32(slice1, slice2 []uint32) []uint32 {
	m := make(map[uint32]int)
	nn := make([]uint32, 0)
	inter := SliceInterI32(slice1, slice2)
	for _, v := range inter {
		m[v]++
	}
	for _, v := range slice1 {
		times, _ := m[v]
		if times == 0 {
			nn = append(nn, v)
		}
	}
	return nn
}
func SliceDiffI64(slice1, slice2 []int64) []int64 {
	m := make(map[int64]int)
	nn := make([]int64, 0)
	inter := SliceInterI64(slice1, slice2)
	for _, v := range inter {
		m[v]++
	}
	for _, v := range slice1 {
		times, _ := m[v]
		if times == 0 {
			nn = append(nn, v)
		}
	}
	return nn
}
func SliceDiffStr(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	inter := SliceInterStr(slice1, slice2)
	for _, v := range inter {
		m[v]++
	}
	for _, v := range slice1 {
		times, _ := m[v]
		if times == 0 {
			nn = append(nn, v)
		}
	}
	return nn
}

// /////////////////////////////////////////////////
// 拆分
func SliceSplitI32(slice []uint32, start, end int) (src, des []uint32) {
	if len(slice) == 0 {
		return
	}
	if end > len(slice) {
		end = len(slice)
	}
	if start < 0 {
		start = 0
	}
	des = append(des, slice[start:end]...)

	src = append(src, slice[0:start]...)
	src = append(src, slice[end:(len(slice))]...)
	return
}
func SliceSplitI64(slice []int64, start, end int) (src, des []int64) {
	if len(slice) == 0 {
		return
	}
	if end > len(slice) {
		end = len(slice)
	}
	if start < 0 {
		start = 0
	}
	des = append(des, slice[start:end]...)

	src = append(src, slice[0:start]...)
	src = append(src, slice[end:(len(slice))]...)
	return
}
func SliceSplitStr(slice []string, start, end int) (src, des []string) {
	if len(slice) == 0 {
		return
	}
	if end > len(slice) {
		end = len(slice)
	}
	if start < 0 {
		start = 0
	}
	des = append(des, slice[start:end]...)

	src = append(src, slice[0:start]...)
	src = append(src, slice[end:(len(slice))]...)
	return
}

// /////////////////////////////////////////////////
// 去重
func SliceUniqueI32(slice [][]uint32) [][]uint32 {
	result := make([][]uint32, 0)
	for i := 0; i < len(slice); i++ {
		flag := false
		for j := i + 1; j < len(slice); j++ {
			if reflect.DeepEqual(slice[i], slice[j]) {
				flag = true
				break
			}
		}
		if !flag {
			result = append(result, slice[i])
		}
	}
	return result
}
func SliceUniqueI64(slice [][]int64) [][]int64 {
	result := make([][]int64, 0)
	for i := 0; i < len(slice); i++ {
		flag := false
		for j := i + 1; j < len(slice); j++ {
			if reflect.DeepEqual(slice[i], slice[j]) {
				flag = true
				break
			}
		}
		if !flag {
			result = append(result, slice[i])
		}
	}
	return result
}
func SliceUniqueStr(slice [][]string) [][]string {
	result := make([][]string, 0)
	for i := 0; i < len(slice); i++ {
		flag := false
		for j := i + 1; j < len(slice); j++ {
			if reflect.DeepEqual(slice[i], slice[j]) {
				flag = true
				break
			}
		}
		if !flag {
			result = append(result, slice[i])
		}
	}
	return result
}

// /////////////////////////////////////////////////
// 排列组合 递归实现 n个元素中取m个元素的所有组合
// 例:{1,2,3} 取 2, 输出{{1,2}, {2,3}, {1,3}}
// 例:{1,2,3} 取 1, 输出{{1}, {2}, {3}}
func SliceComputeI32(slice []uint32, num int) [][]uint32 {
	result := [][]uint32{}
	if len(slice) == num {
		result = append(result, slice)
	} else {
		makeCount := len(slice) - num + 1
		for i := 0; i < makeCount; i++ {
			if num > 1 {
				split, _ := SliceSplitI32(slice, 0, i+1)
				childCm := SliceComputeI32(split, num-1)
				for _, v := range childCm {
					v = SliceInsertI32(v, slice[i], 0)
					result = append(result, v)
				}
			} else if num == 1 {
				result = append(result, []uint32{slice[i]})
			}
		}
	}
	return SliceUniqueI32(result)
}
func SliceComputeI64(slice []int64, num int) [][]int64 {
	result := [][]int64{}
	if len(slice) == num {
		result = append(result, slice)
	} else {
		makeCount := len(slice) - num + 1
		for i := 0; i < makeCount; i++ {
			if num > 1 {
				split, _ := SliceSplitI64(slice, 0, i+1)
				childCm := SliceComputeI64(split, num-1)
				for _, v := range childCm {
					v = SliceInsertI64(v, slice[i], 0)
					result = append(result, v)
				}
			} else if num == 1 {
				result = append(result, []int64{slice[i]})
			}
		}
	}
	return SliceUniqueI64(result)
}
func SliceComputeStr(slice []string, num int) [][]string {
	result := [][]string{}
	if len(slice) == num {
		result = append(result, slice)
	} else {
		makeCount := len(slice) - num + 1
		for i := 0; i < makeCount; i++ {
			if num > 1 {
				split, _ := SliceSplitStr(slice, 0, i+1)
				childCm := SliceComputeStr(split, num-1)
				for _, v := range childCm {
					v = SliceInsertStr(v, slice[i], 0)
					result = append(result, v)
				}
			} else if num == 1 {
				result = append(result, []string{slice[i]})
			}
		}
	}
	return SliceUniqueStr(result)
}

// /////////////////////////////////////////////////
// 排列组合 递归实现 n个元素中取最多m个元素的所有组合
// 例:{1,2,3} 取 2, 输出{{1}, {2}, {3}, {1,2}, {2,3}, {1,3}}
func SliceComputeAllI32(slice []uint32, num int) [][]uint32 {
	result := [][]uint32{}
	for i := 1; i <= num; i++ {
		result = append(result, SliceComputeI32(slice, i)...)
	}
	return result
}
func SliceComputeAllI64(slice []int64, num int) [][]int64 {
	result := [][]int64{}
	for i := 1; i <= num; i++ {
		result = append(result, SliceComputeI64(slice, i)...)
	}
	return result
}
func SliceComputeAllStr(slice []string, num int) [][]string {
	result := [][]string{}
	for i := 1; i <= num; i++ {
		result = append(result, SliceComputeStr(slice, i)...)
	}
	return result
}

// /////////////////////////////////////////////////
// 笛卡尔积算法 多个数组的排列组合
// 例:{{1,2,3}, {4,5}}, 输出{{1,4}, {1,5}, {2,4}, {2,5}, {3,4}, {3,5}}
func _SlicePermutationsI32(slices [][]uint32, out *[][]uint32, idx int, idxes []int) {
	if idx < len(slices) {
		for i := 0; i < len(slices[idx]); i++ {
			idxes[idx] = i
			_SlicePermutationsI32(slices, out, idx+1, idxes)
		}
	} else {
		cm := []uint32{}
		for i := 0; i < len(slices); i++ {
			cm = append(cm, slices[i][idxes[i]])
		}
		*out = append(*out, cm)
	}
}
func SlicePermutationsI32(slices [][]uint32) [][]uint32 {
	result := [][]uint32{}
	idxes := make([]int, len(slices))
	_SlicePermutationsI32(slices, &result, 0, idxes)
	return result
}
func _SlicePermutationsI64(slices [][]int64, out *[][]int64, idx int, idxes []int) {
	if idx < len(slices) {
		for i := 0; i < len(slices[idx]); i++ {
			idxes[idx] = i
			_SlicePermutationsI64(slices, out, idx+1, idxes)
		}
	} else {
		cm := []int64{}
		for i := 0; i < len(slices); i++ {
			cm = append(cm, slices[i][idxes[i]])
		}
		*out = append(*out, cm)
	}
}
func SlicePermutationsI64(slices [][]int64) [][]int64 {
	result := [][]int64{}
	idxes := make([]int, len(slices))
	_SlicePermutationsI64(slices, &result, 0, idxes)
	return result
}
func _SlicePermutationsStr(slices [][]string, out *[][]string, idx int, idxes []int) {
	if idx < len(slices) {
		for i := 0; i < len(slices[idx]); i++ {
			idxes[idx] = i
			_SlicePermutationsStr(slices, out, idx+1, idxes)
		}
	} else {
		cm := []string{}
		for i := 0; i < len(slices); i++ {
			cm = append(cm, slices[i][idxes[i]])
		}
		*out = append(*out, cm)
	}
}
func SlicePermutationsStr(slices [][]string) [][]string {
	result := [][]string{}
	idxes := make([]int, len(slices))
	_SlicePermutationsStr(slices, &result, 0, idxes)
	return result
}
