package utils

import (
	"log/slog"
	"net"
	"strings"
)

func GetIp() (ip string) {
	ips, err := net.InterfaceAddrs()
	if err != nil {
		slog.Error("get interface addresses failed", "error", err)
		return ip
	}

	for _, a := range ips {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ip = ipNet.IP.String()
				if strings.HasPrefix(ip, "10") {
					return
				}
				if strings.HasPrefix(ip, "172") {
					return
				}
				if strings.HasPrefix(ip, "192.168") {
					return
				}
				ip = ""
			}
		}
	}
	return
}
