package resolver

import (
	"testing"
	"io/ioutil"
	"os"
	"path/filepath"
	"fmt"
	"github.com/miekg/dns"
	"net"
	"time"
	"sync"
)
const normal string = `
# Comment
domain somedomain.com
nameserver 10.28.10.2
nameserver 11.28.10.1
`

const abnormal string = `
nameserver asvcd
`

const missingNewline string = `
domain somedomain.com
nameserver 10.28.10.2
nameserver 11.28.10.1` // <- NOTE: NO newline.

func TestNameserver(t *testing.T)          { testNewResolver(t, normal) }
func TestMissingFinalNewLine(t *testing.T) { testNewResolver(t, missingNewline) }
func testNewResolver(t *testing.T, data string) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	path := filepath.Join(tempDir, "resolv.conf")
	if err := ioutil.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	r, err := NewResolver(path, 200, true)

	if err != nil {
		t.Errorf("error new resolver: %v", err)
	}
	if l := len(r.config.Servers); l != 2 {
		t.Errorf("incorrect number of nameservers detected: %d", l)
	}
	if l := len(r.config.Search); l != 1 {
		t.Errorf("domain directive not parsed correctly: %v", r.config.Search)
	} else {
		if r.config.Search[0] != "somedomain.com" {
			t.Errorf("domain is unexpected: %v", r.config.Search[0])
		}
	}
}

func TestNewResolver_normal(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	path := filepath.Join(tempDir, "resolv.conf")
	if err := ioutil.WriteFile(path, []byte(normal), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	r, err := NewResolver(path, 200, true)

	if err != nil {
		t.Errorf("error new resolver: %v", err)
	}
	if l := len(r.config.Servers); l != 2 {
		t.Errorf("incorrect number of nameservers detected: %d", l)
	}
	if l := len(r.config.Search); l != 1 {
		t.Errorf("domain directive not parsed correctly: %v", r.config.Search)
	} else {
		if r.config.Search[0] != "somedomain.com" {
			t.Errorf("domain is unexpected: %v", r.config.Search[0])
		}
	}
}

func TestNewResolver_abnormal(t *testing.T) {

	_, err := NewResolver("abc", 200, true)

	if err == nil {
		t.Errorf("error testing new resolver abnormal process: %v", err)
	}


}

func TestResolver_Nameservers(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	path := filepath.Join(tempDir, "resolv.conf")
	if err := ioutil.WriteFile(path, []byte(normal), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	r, err := NewResolver(path, 200, true)

	servers := r.Nameservers()
	if l := len(servers); l != 2 {
		t.Errorf("incorrect number of nameservers detected: %d", l)
	}
	for _, s := range servers{
		fmt.Printf("server: %s\n", s)
	}

}

func TestUdpResolver_Lookup_normal(t *testing.T) {
	dns.HandleFunc("example.com.", HelloServer)
	defer dns.HandleRemove("example.com.")
	s, addr, err := RunLocalUDPServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}
	host, port, err := net.SplitHostPort(addr)
	if nil != err {
		t.Fatalf("unable to run test server %v", err)
	}
	var resolveconf string = "nameserver " +  host + "#" + port


	defer s.Shutdown()


	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)


	path := filepath.Join(tempDir, "resolv.conf")
	if err := ioutil.WriteFile(path, []byte(resolveconf), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}

	re, err := NewResolver(path, 200, true)

	if err != nil {
		t.Errorf("error parsing resolv.conf: %v", err)
	}

	servers := re.Nameservers()
	if l := len(servers); l != 1 {
		t.Errorf("incorrect number of nameservers detected: %d", l)
	}

	for _, s := range servers{
		fmt.Printf("server: %s\n", s)
	}

	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeSOA)

	r, err :=re.Lookup("udp", m)

	if err != nil {
		t.Errorf("failed to exchange: %v", err)
	}
	if r == nil {
		t.Errorf("failed to get an valid answer\n%v", r)
	}
	if r.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", r)
	}
}

func TestTcpResolver_Lookup_normal(t *testing.T) {
	dns.HandleFunc("example.com.", HelloServer)
	defer dns.HandleRemove("example.com.")

	// This uses TCP just to make it slightly different than TestClientSync
	s, addrstr, err := RunLocalTCPServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}

	host, port, err := net.SplitHostPort(addrstr)
	if nil != err {
		t.Fatalf("unable to run test server %v", err)
	}
	var resolveconf string = "nameserver " +  host + "#" + port

	defer s.Shutdown()


	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)


	path := filepath.Join(tempDir, "resolv.conf")
	if err := ioutil.WriteFile(path, []byte(resolveconf), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}

	re, err := NewResolver(path, 200, true)

	if err != nil {
		t.Errorf("error parsing resolv.conf: %v", err)
	}

	servers := re.Nameservers()
	if l := len(servers); l != 1 {
		t.Errorf("incorrect number of nameservers detected: %d", l)
	}

	for _, s := range servers{
		fmt.Printf("server: %s\n", s)
	}

	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeSOA)

	r, err :=re.Lookup("tcp", m)

	if err != nil {
		t.Errorf("failed to exchange: %v", err)
	}
	if r == nil {
		t.Errorf("failed to get an valid answer\n%v", r)
	}
	if r.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", r)
	}
}

func TestResolver_Lookup_abnormal(t *testing.T) {

	dns.HandleFunc("example.com.", HelloServer)
	defer dns.HandleRemove("example.com.")
	s, addr, err := RunLocalUDPServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}
	host, port, err := net.SplitHostPort(addr)
	if nil != err {
		t.Fatalf("unable to run test server %v", err)
	}
	var resolveconf string = "nameserver " +  host + "#" + port


	defer s.Shutdown()


	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)


	path := filepath.Join(tempDir, "resolv.conf")
	if err := ioutil.WriteFile(path, []byte(resolveconf), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}

	re, err := NewResolver(path, 200, true)

	if err != nil {
		t.Errorf("error parsing resolv.conf: %v", err)
	}

	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeSOA)

	_, err =re.Lookup("abc", m)

	if err == nil {
		t.Errorf("error testing lookup abnormal process: %v", err)
	}
	t.Logf("expected error: %v", err)


}

func HelloServer(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)

	m.Extra = make([]dns.RR, 1)
	m.Extra[0] = &dns.TXT{Hdr: dns.RR_Header{Name: m.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}, Txt: []string{"Hello world"}}
	w.WriteMsg(m)
}

func RunLocalUDPServer(laddr string) (*dns.Server, string, error) {
	server, l, _, err := RunLocalUDPServerWithFinChan(laddr)

	return server, l, err
}

func RunLocalUDPServerWithFinChan(laddr string) (*dns.Server, string, chan struct{}, error) {
	pc, err := net.ListenPacket("udp", laddr)
	if err != nil {
		return nil, "", nil, err
	}
	server := &dns.Server{PacketConn: pc, ReadTimeout: time.Hour, WriteTimeout: time.Hour}

	waitLock := sync.Mutex{}
	waitLock.Lock()
	server.NotifyStartedFunc = waitLock.Unlock

	fin := make(chan struct{}, 0)

	go func() {
		server.ActivateAndServe()
		close(fin)
		pc.Close()
	}()

	waitLock.Lock()
	return server, pc.LocalAddr().String(), fin, nil
}

func RunLocalTCPServer(laddr string) (*dns.Server, string, error) {
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		return nil, "", err
	}

	server := &dns.Server{Listener: l, ReadTimeout: time.Hour, WriteTimeout: time.Hour}

	waitLock := sync.Mutex{}
	waitLock.Lock()
	server.NotifyStartedFunc = waitLock.Unlock

	go func() {
		server.ActivateAndServe()
		l.Close()
	}()

	waitLock.Lock()
	return server, l.Addr().String(), nil
}


