// Package avoid implements a plugin.
/*
 * This code comes from the coredns repo: https://github.com/coredns/demo
 * under apache 2 license.
 */
package avoid

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

// Avoid is a plugin in CoreDNS
type Avoid struct{}

// ServeDNS implements the plugin.Handler interface.
func (p Avoid) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.Name()

	reply := "8.8.8.8"
	if strings.HasPrefix(state.IP(), "172.") || strings.HasPrefix(state.IP(), "127.") {
		reply = "1.1.1.1"
	}
	fmt.Printf("Received query %s from %s, expected to reply %s\n", qname, state.IP(), reply)

	answers := []dns.RR{}

	if state.QType() != dns.TypeA {
		return dns.RcodeNameError, nil
	}

	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeA, Class: dns.ClassINET}
	rr.A = net.ParseIP(reply).To4()

	answers = append(answers, rr)

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = answers

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (p Avoid) Name() string { return "avoid" }
