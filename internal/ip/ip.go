package ip

import (
	"github.com/vishvananda/netlink"
)

type IP struct {
	addr *netlink.Addr
	link netlink.Link
}

func NewIP(ip string, device string) (i *IP, err error) {
	addr, err := netlink.ParseAddr(ip)
	if err != nil {
		return
	}
	link, err := netlink.LinkByAlias(device)
	if err != nil {
		return
	}
	return &IP{
		addr: addr,
		link: link,
	}, nil
}

func (ip *IP) Bind() (err error) {
	return netlink.AddrAdd(ip.link, ip.addr)
}

func (ip *IP) Unbind() (err error) {
	return netlink.AddrDel(ip.link, ip.addr)
}
