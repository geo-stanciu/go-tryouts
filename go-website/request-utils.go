package main

import (
	"net"
	"net/http"
	"strings"
)

func getClientIP(r *http.Request) string {
	ips := r.Header.Get("X-Forwarded-For")

	ipList := strings.Split(ips, ", ")

	if len(ipList) > 0 && len(ipList[0]) > 0 {
		return ipList[0]
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	return ip
}

func isRequestFromLocalhost(r *http.Request) bool {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	if ip == "127.0.0.1" || ip == "::1" {
		return true
	}

	return false
}
