package utils

import "unsafe"

func FastBoolToInt(b bool) int {
	// found on https://dev.to/chigbeef_77/bool-int-but-stupid-in-go-3jb3 (way 7)
	// based on the quake fast inverse sqrt
	// first make an unsafe (does not hold type info) pointer to the bool
	// cast that pointer to a pointer to a byte
	// (this is faster than casting to int for some reason)
	// finally, dereference and cast to integer (will be 1 or 0)
	return int(*(*byte)(unsafe.Pointer(&b)))
}