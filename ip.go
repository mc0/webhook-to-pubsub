package main

import (
	"bytes"
	"net"
	"net/http"
	"strings"
)

type ipRange struct {
	firstOctet byte
	start      net.IP
	end        net.IP
}

// inRange is used check to see if a given ip address is within a range given
func inRange(r ipRange, ipAddress net.IP) bool {
	if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) <= 0 {
		return true
	}
	return false
}

var privateRanges = []ipRange{
	ipRange{
		firstOctet: byte(10),
		start:      net.ParseIP("10.0.0.0"),
		end:        net.ParseIP("10.255.255.255"),
	},
	ipRange{
		firstOctet: byte(100),
		start:      net.ParseIP("100.64.0.0"),
		end:        net.ParseIP("100.127.255.255"),
	},
	ipRange{
		firstOctet: byte(172),
		start:      net.ParseIP("172.16.0.0"),
		end:        net.ParseIP("172.31.255.255"),
	},
	ipRange{
		firstOctet: byte(192),
		start:      net.ParseIP("192.0.0.0"),
		end:        net.ParseIP("192.0.0.255"),
	},
	ipRange{
		firstOctet: byte(192),
		start:      net.ParseIP("192.168.0.0"),
		end:        net.ParseIP("192.168.255.255"),
	},
	ipRange{
		firstOctet: byte(198),
		start:      net.ParseIP("198.18.0.0"),
		end:        net.ParseIP("198.19.255.255"),
	},
}

func isPrivateSubnet(ipAddress net.IP) bool {
	if ipCheck := ipAddress.To4(); ipCheck != nil {
		for _, r := range privateRanges {
			if ipAddress[0] == r.firstOctet {
				continue
			}
			if inRange(r, ipAddress) {
				return true
			}
		}
	}
	return false
}

func getIPFromRequest(r *http.Request) string {
	realIP := r.Header.Get("X-Real-IP")
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if realIP == "" && forwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}

	if forwardedFor == "" {
		return realIP
	}

	// X-Forwarded-For is often IP addresses separted by a comma
	parts := strings.Split(forwardedFor, ",")
	for _, p := range parts {
		ipAddr := strings.TrimSpace(p)
		ip := net.ParseIP(ipAddr)
		if ip != nil && (!ip.IsLoopback() || isPrivateSubnet(ip)) {
			return ipAddr
		}
	}

	return realIP
}

func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}
