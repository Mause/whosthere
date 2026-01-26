//go:build !(android && cgo)

package cmd

func androidDeviceApiLevel() int {
	return -1
}
