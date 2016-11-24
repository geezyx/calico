// Copyright (c) 2016 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The ip package contains yet another IP address (and CIDR) type :-).   The
// types differ from the ones in the net package in that they are backed by
// fixed-sized arrays of the appropriate size.  The key advantage of
// using a fixed-size array is that it makes the types hashable so they can
// be used as map keys.  In addition, they can be converted to net.IP by
// slicing.
package ip

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	calinet "github.com/projectcalico/libcalico-go/lib/net"
	"net"
)

// Addr represents either an IPv4 or IPv6 IP address.
type Addr interface {
	// Version returns the IP version; 4 or 6.
	Version() uint8
	// AsNetIP returns a net.IP, which is backed by/shares storage with
	// this object.
	AsNetIP() net.IP
	AsCalicoNetIP() calinet.IP
	String() string
}

type V4Addr [4]byte

func (a V4Addr) Version() uint8 {
	return 4
}

func (a V4Addr) AsNetIP() net.IP {
	return net.IP(a[0:net.IPv4len])
}

func (a V4Addr) AsCalicoNetIP() calinet.IP {
	return calinet.IP{a.AsNetIP()}
}

func (a V4Addr) String() string {
	return a.AsNetIP().String()
}

type V6Addr [16]byte

func (a V6Addr) Version() uint8 {
	return 6
}

func (a V6Addr) AsNetIP() net.IP {
	return net.IP(a[0:net.IPv6len])
}

func (a V6Addr) AsCalicoNetIP() calinet.IP {
	return calinet.IP{a.AsNetIP()}
}

func (a V6Addr) String() string {
	return a.AsNetIP().String()
}

type CIDR interface {
	Version() uint8
	Addr() Addr
	Prefix() uint8
	String() string
	ToIPNet() net.IPNet
}

type V4CIDR struct {
	addr   V4Addr
	prefix uint8
}

func (c V4CIDR) Version() uint8 {
	return 4
}

func (c V4CIDR) Addr() Addr {
	return c.addr
}

func (c V4CIDR) Prefix() uint8 {
	return c.prefix
}

func (c V4CIDR) ToIPNet() net.IPNet {
	return net.IPNet{
		IP:   c.Addr().AsNetIP(),
		Mask: net.CIDRMask(int(c.Prefix()), 32),
	}
}

func (c V4CIDR) String() string {
	return fmt.Sprintf("%s/%v", c.addr.String(), c.prefix)
}

type V6CIDR struct {
	addr   V6Addr
	prefix uint8
}

func (c V6CIDR) Version() uint8 {
	return 6
}

func (c V6CIDR) Addr() Addr {
	return c.addr
}

func (c V6CIDR) Prefix() uint8 {
	return c.prefix
}

func (c V6CIDR) ToIPNet() net.IPNet {
	return net.IPNet{
		IP:   c.Addr().AsNetIP(),
		Mask: net.CIDRMask(int(c.Prefix()), 128),
	}
}

func (c V6CIDR) String() string {
	return fmt.Sprintf("%s/%v", c.addr.String(), c.prefix)
}

func FromNetIP(netIP net.IP) Addr {
	if len(netIP) == 4 {
		ip := V4Addr{}
		for ii, b := range netIP {
			ip[ii] = b
		}
		return ip
	} else {
		ip := V6Addr{}
		for ii, b := range netIP {
			ip[ii] = b
		}
		return ip
	}
}

func CIDRFromIPNet(ipNet calinet.IPNet) CIDR {
	ones, _ := ipNet.Mask.Size()
	ip := FromNetIP(ipNet.IP)
	if ip.Version() == 4 {
		return V4CIDR{
			addr:   ip.(V4Addr),
			prefix: uint8(ones),
		}
	} else {
		return V6CIDR{
			addr:   ip.(V6Addr),
			prefix: uint8(ones),
		}
	}
}

func MustParseCIDR(s string) CIDR {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		log.WithError(err).WithField("cidr", s).Panic("Failed to parse CIDR")
	}
	return CIDRFromIPNet(calinet.IPNet{*ipNet})
}
