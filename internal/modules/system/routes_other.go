//go:build !linux

package system

import "fmt"

func defaultRouteInterface() (string, error) {
	return "", fmt.Errorf("default-route lookup is unavailable on this platform")
}
