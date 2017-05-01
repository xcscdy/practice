package resolver

import (
	"github.com/miekg/dns"
	"time"
	"sync"
	"fmt"
	"strings"
	"net"
	"errors"
)

type (
	IResolver interface {
		Lookup(net string, req *dns.Msg) (message *dns.Msg, err error)
	}

	Resolver struct {
		config *dns.ClientConfig
		Interval   uint32
		SetEDNS0   bool
	}
)

type ResolvError struct {
	qname, net  string
	nameservers []string
}

func (e ResolvError) Error() string {
	errmsg := fmt.Sprintf("%s resolv failed on %s (%s)", e.qname, strings.Join(e.nameservers, "; "), e.net)
	return errmsg
}

func NewResolver(resolvconf string, interval uint32, setEDNS0 bool) (*Resolver, error)  {
	cc, err := dns.ClientConfigFromFile(resolvconf)
	if err != nil {
		fmt.Print("error parsing resolv.conf: %v", err)
		return nil, err
	}
	if interval == 0 {
		interval = 200
	}
	r := &Resolver{config:cc, Interval:interval, SetEDNS0:setEDNS0 }

	return r, nil
}


func (r *Resolver)Lookup(net string, req *dns.Msg) (message *dns.Msg, err error)  {

	if nil == req {
		return nil, errors.New("input request is nil")
	}
	c := &dns.Client{
		Net:          net,
		ReadTimeout:  r.Timeout(),
		WriteTimeout: r.Timeout(),
	}

	if net == "udp" && r.SetEDNS0 {
		req = req.SetEdns0(65535, true)
	}

	qname := req.Question[0].Name

	res := make(chan *dns.Msg, 1)
	var wg sync.WaitGroup

	L := func(nameserver string) {
		defer wg.Done()
		r, rtt, err := c.Exchange(req, nameserver)
		if err != nil {
			fmt.Sprintln("%s socket error on %s", qname, nameserver)
			fmt.Sprintln("error:%s", err.Error())
			return
		}
		// If SERVFAIL happen, should return immediately and try another upstream resolver.
		// However, other Error code like NXDOMAIN is an clear response stating
		// that it has been verified no such domain existas and ask other resolvers
		// would make no sense. See more about #20
		if r != nil && r.Rcode != dns.RcodeSuccess {
			fmt.Sprintln("%s failed to get an valid answer on %s", qname, nameserver)
			if r.Rcode == dns.RcodeServerFailure {
				return
			}
		} else {
			fmt.Sprintln("%s resolv on %s (%s) ttl: %d", UnFqdn(qname), nameserver, net, rtt)
		}
		select {
		case res <- r:
		default:
		}
	}

	ticker := time.NewTicker(time.Duration(r.Interval) * time.Millisecond)
	defer ticker.Stop()
	// Start lookup on each nameserver top-down, in every second
	for _, nameserver := range r.Nameservers() {
		wg.Add(1)
		go L(nameserver)
		// but exit early, if we have an answer
		select {
		case r := <-res:
			return r, nil
		case <-ticker.C:
			continue
		}
	}
	// wait for all the namservers to finish
	wg.Wait()
	select {
	case r := <-res:
		return r, nil
	default:
		return nil, ResolvError{qname, net, r.Nameservers()}
	}
}

// Namservers return the array of nameservers, with port number appended.
// '#' in the name is treated as port separator, as with dnsmasq.
func (r *Resolver) Nameservers() (ns []string) {
	for _, server := range r.config.Servers {
		if i := strings.IndexByte(server, '#'); i > 0 {
			server = net.JoinHostPort(server[:i], server[i+1:])
		} else {
			server = net.JoinHostPort(server, r.config.Port)
		}
		ns = append(ns, server)
	}
	return
}

func (r *Resolver) Timeout() time.Duration {
	return time.Duration(r.config.Timeout) * time.Second
}

func UnFqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}
