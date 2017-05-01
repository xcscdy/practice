package middleware

import (
	"testing"
	"github.com/miekg/dns"
	"fmt"
)

type (
	middlewareA struct {

	}
	middlewareB struct {

	}
	middlewareC struct {

	}
)

func (a *middlewareA) DoServer(w dns.ResponseWriter, r *dns.Msg, m *MiddlewareChain)  {
	fmt.Println("in middleware A DoServer")
	m.DoServer(w,r)

}

func (b *middlewareB) DoServer(w dns.ResponseWriter, r *dns.Msg, m *MiddlewareChain)  {
	fmt.Println("in middleware B DoServer")
	m.DoServer(w,r)
}

func (c *middlewareC) DoServer(w dns.ResponseWriter, r *dns.Msg, m *MiddlewareChain)  {
	fmt.Println("in middleware C DoServer")
	m.DoServer(w,r)
}

func TestMiddlewareChain_DoServer(t *testing.T) {
	mn := middlewareNode{midllerware: &middlewareA{}}
	mc := MiddlewareChain{currentNode: &mn}
	mc.DoServer(nil, nil)
}

func TestMiddlewareContainer_DoServer(t *testing.T) {
	mcontainer := MiddlewareContainer{}

	mcontainer.AddMiddleware(&middlewareA{})
	mcontainer.AddMiddleware(&middlewareB{})
	mcontainer.AddMiddleware(&middlewareC{})

	mcontainer.DoServer(nil,nil)
	mcontainer.DoServer(nil,nil)
	mcontainer.DoServer(nil,nil)

}
