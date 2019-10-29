package router

import (
	"net"
	"fmt"

	"github.com/vishvananda/netlink"
)

// Gateway interfaces to router
type Gateway interface {
	Table() *routes
}

// NextHop represents nexthop / gateway
type NextHop struct {
	IP net.IP
}

// Router represents router
type Router struct {
	routes *routes
}

// Table returns touting table
func (r *Router) Table() *routes {
	return r.routes
}

// Route represents a route
type Route struct {
	NextHop NextHop
	Dst     *net.IPNet
}

// routes represents array of routes
type routes struct {
	table []Route
}

// Add appends a new route to table and operating system
func (r *routes) Add(dst *net.IPNet, nexthop net.IP) error {
	// check if route exist
	for _, route := range r.table {
		if route.Dst == dst {
			return fmt.Errorf("route exist %s %s", dst, nexthop)
		}
	}

	r.table = append(r.table, Route{
		NextHop: NextHop{IP:nexthop},
		Dst: dst,
	})	

	// add route to operating system 
	ifce, err := netlink.LinkByName("radvpn")
	if err != nil {
		return err
	}

	rr := &netlink.Route{
		Dst: dst,
		LinkIndex: ifce.Attrs().Index,
	}
	err = netlink.RouteAdd(rr)	
	if err != nil {
		return err
	}

	return nil
}

// Get returns nexthop for a specific dest.
func (r routes) Get(dst net.IP) net.IP {
	for _, route := range r.table {
		if route.Dst.Contains(dst) {
			return route.NextHop.IP
		}
	}

	return nil
}

// Dump prints out all routing table
func (r routes) Dump() {
	fmt.Println("destination\tnexthop")
	for _, route := range r.table {
		fmt.Println(route.Dst, route.NextHop.IP)
	}	
}

// New constructs a new router
func New() *Router {
	return &Router{new(routes)}
}
