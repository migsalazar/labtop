//go:build linux

package system

import "os"

func defaultRouteInterface() (string, error) {
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return "", err
	}
	defer file.Close()
	return parseDefaultRoute(file)
}
