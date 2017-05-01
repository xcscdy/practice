package middleware

import (
	"github.com/miekg/dns"
)

type(
	Middleware interface {
		DoServer(dns.ResponseWriter, *dns.Msg, *MiddlewareChain)
	}
	MiddlewareChain struct {
		currentNode *middlewareNode
	}

	middlewareNode struct {
		midllerware Middleware
		nextNode *middlewareNode
	}
	MiddlewareContainer struct {
		head *middlewareNode
		tail *middlewareNode
		middlewares []middlewareNode
	}
)

func (hc *MiddlewareChain) DoServer(writer dns.ResponseWriter, msg *dns.Msg) {
	if nil == hc.currentNode {
		return
	} else {
		h := hc.currentNode.midllerware
		hc.currentNode = hc.currentNode.nextNode
		h.DoServer(writer, msg, hc)
	}
}

func (c *MiddlewareContainer) AddMiddleware(m Middleware)  {
	if nil == m {
		return
	}
	node := middlewareNode{midllerware:m}
	c.middlewares = append(c.middlewares, node)
	if nil == c.head {
		c.head = &node
		c.tail = &node
		return
	}
	c.tail.nextNode = &node
	c.tail = &node
}

func (c *MiddlewareContainer) DoServer(writer dns.ResponseWriter, msg *dns.Msg){

	if nil == c.head {
		return
	}
	chain := &MiddlewareChain{currentNode: c.head}
	chain.DoServer(writer, msg)
}

