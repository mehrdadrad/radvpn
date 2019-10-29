package router

import (
	"net"
	"fmt"
)

type Gateway interface {
	Table() *routes
}

type NextHop struct {
	IP net.IP
}

type Router struct {
	routes *routes
}

func (r *Router) Table() *routes {
	return r.routes
}

type Route struct {
	NextHop NextHop
	Dst     *net.IPNet
}

type routes struct {
	table []Route
}

func (r *routes) Add(dst *net.IPNet, nexthop net.IP) {
	// TODO check duplicate
	r.table = append(r.table, Route{
		NextHop: NextHop{IP:nexthop},
		Dst: dst,
	})	
}

func (r routes) Get(dst net.IP) net.IP {
	for _, route := range r.table {
		if route.Dst.Contains(dst) {
			return route.NextHop.IP
		}
	}

	return nil
}

func (r routes) Dump() {
	for _, route := range r.table {
		fmt.Println(route.Dst, route.NextHop.IP)
	}	
}

func New() *Router {
	return &Router{new(routes)}
}
