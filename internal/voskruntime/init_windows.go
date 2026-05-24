//go:build windows && cgo

package voskruntime

func init() {
	if err := Ensure(); err != nil {
		panic("vosk runtime: " + err.Error())
	}
}
