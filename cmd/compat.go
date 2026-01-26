//go:build android && cgo

package cmd

/*
#cgo LDFLAGS: -landroid
// #include <android/api-level.h>
// #include "cgo_helpers.h"
*/
import "C"

func getApiLevel() int {
	level := int(C.android_get_device_api_level())
	return int(level)
}
