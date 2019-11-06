package router

import (
	"context"
	"net"
	"testing"
)

func TestAddFromRouter(t *testing.T) {
	r := New(context.Background()).Table()

	subnetStr := "10.0.1.0/24"
	nexthopStr := "192.168.55.1"

	_, subnet, _ := net.ParseCIDR(subnetStr)
	nexthop := net.ParseIP(nexthopStr)

	r.addToRouter(subnet, nexthop)

	if len(r.table) == 1 {
		if r.table[0].NextHop.IP.String() != nexthopStr {
			t.Errorf("expect nexthop %s but got, %s", nexthopStr, r.table[0].NextHop.IP.String())
		}

		if r.table[0].NetworkID.String() != subnetStr {
			t.Errorf("expect networkid %s but got, %s", subnetStr, r.table[0].NetworkID.String())
		}
	} else {
		t.Error("route.add method can't add a route")
	}
}

func TestDelFromRouter(t *testing.T) {
	r := New(context.Background()).Table()

	subnetStr := "10.0.1.0/24"
	nexthopStr := "192.168.55.1"

	_, subnet, _ := net.ParseCIDR(subnetStr)
	nexthop := net.ParseIP(nexthopStr)

	r.addToRouter(subnet, nexthop)

	if len(r.table) == 1 {
		err := r.delFromRouter(subnet, nexthop)
		if err != nil {
			t.Error("unexpected error happened:", err)
		}

		if len(r.table) != 0 {
			t.Error("route.del method can't del a route")
		}
	}
}
