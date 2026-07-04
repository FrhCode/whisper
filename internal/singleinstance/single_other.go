//go:build !windows

package singleinstance

func Lock(string) (func(), error) { return func() {}, nil }
