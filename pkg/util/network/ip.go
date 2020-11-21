package network

import (
	"errors"
	"github.com/pjoc-team/tracing/logger"
	"net"
	"strings"
)

// ErrNotFoundExternalIP not found external ip error
var ErrNotFoundExternalIP = errors.New("not found external ip")

// GetPortByListenAddr get host of listen addr
func GetPortByListenAddr(addr string) string {
	split := strings.Split(addr, ":")
	return split[1]
}


// GetHostIP get local ip
func GetHostIP() (string, error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		logger.Log().Errorf("failed to get ip, error: %v", err.Error())
		return "", err
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}

		}
	}
	return "", ErrNotFoundExternalIP
}
