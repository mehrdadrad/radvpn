package router

import (
	"fmt"
	"net"
	"sync"

	"github.com/vishvananda/netlink"
)

// Gateway interfaces to router
type Gateway interface {
	Table() *Routes
}

// NextHop represents nexthop / gateway
type NextHop struct {
	IP net.IP
}

// Router represents router
type Router struct {
	routes *Routes
}

// Table returns touting table
func (r *Router) Table() *Routes {
	return r.routes
}

// Route represents a route
type Route struct {
	NextHop   NextHop
	NetworkID *net.IPNet
}

// Routes represents array of routes
type Routes struct {
	sync.Mutex

	table []Route
}

// Add appends a new route to table and operating system
func (r *Routes) Add(networkid *net.IPNet, nexthop net.IP) error {
	// check if route exist
	for _, route := range r.table {
		if nexthop.Equal(route.NextHop.IP) && networkid.String() == route.NetworkID.String() {
			return fmt.Errorf("route exist %s %s", networkid, nexthop)
		}
	}

	r.Lock()

	r.table = append(r.table, Route{
		NextHop:   NextHop{IP: nexthop},
		NetworkID: networkid,
	})

	r.Unlock()

	// add route to operating system
	ifce, err := netlink.LinkByName("radvpn")
	if err != nil {
		return err
	}

	route := &netlink.Route{
		Dst:       networkid,
		LinkIndex: ifce.Attrs().Index,
	}
	err = netlink.RouteAdd(route)
	if err != nil {
		return err
	}

	return nil
}

// Get returns nexthop for a specific dest.
func (r *Routes) Get(dst net.IP) net.IP {
	for _, route := range r.table {
		if route.NetworkID.Contains(dst) {
			return route.NextHop.IP
		}
	}

	return nil
}

// Dump prints out all routing table
func (r *Routes) Dump() {
	fmt.Println("networkid\tnexthop")
	for _, route := range r.table {
		fmt.Println(route.NetworkID, route.NextHop.IP)
	}
}

// New constructs a new router
func New() *Router {
	return &Router{new(Routes)}
}
