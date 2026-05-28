//go:build !windows || !cgo

package voskruntime

func Ensure() error { return nil }
