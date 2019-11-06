package router

import (
	"context"
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
	ctx    context.Context
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

func (r *Routes) addToRouter(networkid *net.IPNet, nexthop net.IP) error {
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

	return nil
}

func (r *Routes) addToHost(networkid *net.IPNet, nexthop net.IP) error {
	ifce, err := netlink.LinkByName("radvpn")
	if err != nil {
		return err
	}

	route := &netlink.Route{
		Dst:       networkid,
		LinkIndex: ifce.Attrs().Index,
	}
	return netlink.RouteAdd(route)
}

// Add appends a new route to table and operating system
func (r *Routes) Add(networkid *net.IPNet, nexthop net.IP) error {
	err := r.addToRouter(networkid, nexthop)
	if err != nil {
		return err
	}

	return r.addToHost(networkid, nexthop)
}

func (r *Routes) delFromRouter(networkid *net.IPNet, nexthop net.IP) error {
	var routeRemoved bool

	for k, v := range r.table {
		if v.NetworkID.String() == networkid.String() && nexthop.Equal(v.NextHop.IP) {
			r.Lock()

			r.table[k] = r.table[len(r.table)-1]
			r.table = r.table[:len(r.table)-1]

			r.Unlock()
			routeRemoved = true
			break
		}
	}

	if !routeRemoved {
		return fmt.Errorf("can not delete route, not found %s", networkid.String())
	}

	return nil
}

func (r *Routes) delFromHost(networkid *net.IPNet, nexthop net.IP) error {
	ifce, err := netlink.LinkByName("radvpn")
	if err != nil {
		return err
	}

	route := &netlink.Route{
		Dst:       networkid,
		LinkIndex: ifce.Attrs().Index,
	}

	return netlink.RouteDel(route)
}

// Delete removes a route from table and operating system
func (r *Routes) Delete(networkid *net.IPNet, nexthop net.IP) error {

	err := r.delFromRouter(networkid, nexthop)
	if err != nil {
		return err
	}

	return r.delFromHost(networkid, nexthop)
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
func New(ctx context.Context) *Router {
	return &Router{
		ctx:    ctx,
		routes: new(Routes),
	}
}
