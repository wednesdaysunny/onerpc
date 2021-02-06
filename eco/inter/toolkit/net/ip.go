package net

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

var (
	privateBlocks []*net.IPNet
)

func init() {

	for _, b := range []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		"127.0.0.0/8",
	} {
		if _, block, err := net.ParseCIDR(b); err == nil {
			privateBlocks = append(privateBlocks, block)
		}
	}
}

// LocalIPAddr extracts the public IP address of current machine
func LocalIPAddr() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("Failed to get interface addresses! Err: %v", err)
	}

	var ipAddr []byte

	for _, rawAddr := range addrs {
		var ip net.IP
		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}

		if ip.To4() == nil {
			continue
		}

		if !IsPrivateIpv4(ip.String()) {
			continue
		}

		ipAddr = ip
		break
	}

	if ipAddr == nil {
		return "", fmt.Errorf("No private IP address found, and explicit IP not provided")
	}

	return net.IP(ipAddr).String(), nil
}

// LocalIPAddrWithin picks the IP within the specified subnets.
func LocalIPAddrWithin(subnets []string) (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("Failed to get interface addresses! Err: %v", err)
	}

	var allowSubnets = make([]*net.IPNet, 0)
	for _, subnet := range subnets {
		if _, b, err := net.ParseCIDR(subnet); err == nil {
			allowSubnets = append(allowSubnets, b)
		}
	}

	var ipAddr []byte

	for _, rawAddr := range addrs {
		var ip net.IP
		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}

		if ip.To4() == nil {
			continue
		}

		if !isIPWithin(ip, allowSubnets) {
			continue
		}

		ipAddr = ip
		break
	}

	if ipAddr == nil {
		return "", fmt.Errorf("No private IP address found, and explicit IP not provided")
	}

	return net.IP(ipAddr).String(), nil
}

func isIPWithin(ip net.IP, allowSubnets []*net.IPNet) bool {
	for _, subnet := range allowSubnets {
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

// LocalIPAddrOutside the public IP address of current machine outside the specified private IPs
func LocalIPAddrOutside(privates []string) (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("Failed to get interface addresses! Err: %v", err)
	}

	var ipAddr []byte

	for _, rawAddr := range addrs {
		var ip net.IP
		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}

		if ip.To4() == nil {
			continue
		}

		if !isPrivateIPWithin(ip.String(), privates) {
			continue
		}

		ipAddr = ip
		break
	}

	if ipAddr == nil {
		return "", fmt.Errorf("No private IP address found, and explicit IP not provided")
	}

	return net.IP(ipAddr).String(), nil
}

// IP2Int converts IP to integer.
func IP2Int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func isPrivateIPWithin(ipAddr string, privates []string) bool {
	ip := net.ParseIP(ipAddr)
	var newPrivates []*net.IPNet
	for _, b := range privates {
		if _, block, err := net.ParseCIDR(b); err == nil {
			newPrivates = append(privateBlocks, block)
		}
	}
	for _, priv := range newPrivates {
		if priv.Contains(ip) {
			return true
		}
	}
	return false
}

func isPrivateIP(ipAddr string) bool {
	ip := net.ParseIP(ipAddr)
	for _, priv := range privateBlocks {
		if priv.Contains(ip) {
			return true
		}
	}
	return false
}

func ipHost(ip string) string {
	h := strings.TrimSpace(ip)
	if strings.Index(h, ":") != -1 {
		h, _, _ = net.SplitHostPort(h)
	}
	return h
}

func IsPrivateIpv4(ip string) bool {
	h := ipHost(ip)
	if ipv4 := net.ParseIP(h); ipv4.To4() != nil {
		for _, v := range privateBlocks {
			if v.Contains(ipv4) {
				return true
			}
		}
	}
	return false
}
