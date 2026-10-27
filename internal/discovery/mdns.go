// Package discovery advertises the HTTP server using DNS-SD on local networks.
package discovery

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"

	"dinosaur/internal/config"
	"github.com/hashicorp/mdns"
)

type Advertiser struct{ servers []*mdns.Server }

func Start(c config.Config, id string, logger *slog.Logger) *Advertiser {
	a := &Advertiser{}
	if !c.DiscoveryEnabled {
		return a
	}
	host, rawPort, err := net.SplitHostPort(c.Addr)
	if err != nil {
		return a
	}
	bound := net.ParseIP(host)
	if bound != nil && bound.IsLoopback() || host != "" && bound == nil {
		return a
	}
	port, err := strconv.Atoi(rawPort)
	if err != nil {
		return a
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		logger.Warn("discovery unavailable", "error", err)
		return a
	}
	for _, iface := range interfaces {
		if containerInterface(iface.Name) {
			continue
		}
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagMulticast == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		ips := addresses(iface, bound)
		if len(ips) == 0 {
			continue
		}
		txt := []string{"id=" + id, "name=" + c.ServerName, "api=1"}
		if c.PublicURL != "" && len(c.PublicURL) < 250 {
			txt = append(txt, "url="+c.PublicURL)
		}
		service, err := mdns.NewMDNSService("Dinosaur-"+id[:8], "_dinosaur._tcp", "local.", fmt.Sprintf("dinosaur-%s.local.", id), port, ips, txt)
		if err != nil {
			logger.Warn("discovery service unavailable", "error", err)
			continue
		}
		server, err := mdns.NewServer(&mdns.Config{Zone: service, Iface: &iface, Logger: slog.NewLogLogger(logger.Handler(), slog.LevelWarn)})
		if err != nil {
			logger.Warn("discovery interface unavailable", "interface", iface.Name, "error", err)
			continue
		}
		a.servers = append(a.servers, server)
		logger.Info("LAN discovery enabled", "interface", iface.Name, "port", port)
	}
	return a
}

func containerInterface(name string) bool {
	for _, prefix := range []string{"docker", "veth", "br-", "virbr"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func addresses(iface net.Interface, bound net.IP) []net.IP {
	result := make([]net.IP, 0)
	addrs, err := iface.Addrs()
	if err != nil {
		return result
	}
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil || ip.To4() == nil || !ip.IsPrivate() {
			continue
		}
		if bound != nil && !bound.IsUnspecified() && !bound.Equal(ip) {
			continue
		}
		result = append(result, ip)
	}
	return result
}

func (a *Advertiser) Close() {
	for _, server := range a.servers {
		_ = server.Shutdown()
	}
}
