package main

import (
	"net"

	"github.com/miekg/dns"
)

// injectECS replaces any existing EDNS Client Subnet option on m with subnet.
// subnet must be a valid CIDR; callers are expected to validate beforehand.
func injectECS(m *dns.Msg, subnet string) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return
	}
	ones, _ := ipnet.Mask.Size()

	opt := m.IsEdns0()
	if opt == nil {
		opt = new(dns.OPT)
		opt.Hdr.Name = "."
		opt.Hdr.Rrtype = dns.TypeOPT
		opt.SetUDPSize(dns.DefaultMsgSize)
		m.Extra = append(m.Extra, opt)
	}

	filtered := opt.Option[:0]
	for _, o := range opt.Option {
		if o.Option() != dns.EDNS0SUBNET {
			filtered = append(filtered, o)
		}
	}

	ecs := &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		SourceNetmask: uint8(ones),
		SourceScope:   0,
	}
	if ip4 := ipnet.IP.To4(); ip4 != nil {
		ecs.Family = 1
		ecs.Address = ip4.Mask(ipnet.Mask)
	} else {
		ecs.Family = 2
		ecs.Address = ipnet.IP.Mask(ipnet.Mask)
	}
	opt.Option = append(filtered, ecs)
}
