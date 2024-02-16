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
	"sync"

	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

// TODO: Move from global dict to in-memory storage structure (e.g., memcached or redis)
// TODO: Make assumption that there are multiple authorative servers
var (
	LookupTable = make(map[string]*avoid.DNSEntry{})
	mutex       sync.Mutex
)

// Avoid is a plugin in CoreDNS
type Avoid struct{}

// ServeDNS implements the plugin.Handler interface.
func (p Avoid) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.Name()
	log.Infof("Received query %s from %s\n", qname, state.IP())

	if state.QType() != dns.TypeA || state.QType() != dns.TypeAAAA {
		log.Errorf("invalid request type for this plugin: %v\n", state.QType)
		return dns.RcodeNameError, nil
	}

	// check if the identifier is in our lookup table - for now this is just
	// using the source IP, and we may get into trouble is there is proxy
	// DNS as then we will need to continiously probe the path
	val, ok := LookupTable[state.IP()]
	if ok {
		// we have an entry
	} else {
		// we need to use the default entry

		// TODO: use config file to set default key for deployments
		val2, ok2 := LookupTable[avoid.Default]
		if !ok2 {
			errMsg := fmt.Errorf("Missing default dns entry in table for: %s", avoid.Default)
			log.Error(errMsg)
			return nil, errMsg
		}
		// set value to be the default entry
		val = val2
	}

	// now we need to check what type of entry we are creating and responding to
	// based on the identification we used into the lookup table

	addr, err := net.ParseAddr(state.IP())
	if err != nil {
		log.Errorf("failed to parse incoming requests ip address: %v", err)
		return nil, err
	}

	answers := []dns.RR{}
	if addr.Is4() {
		for record := range val.A {
			rr := new(dns.A)
			rr.Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeA, Class: dns.ClassINET}
			rr.A = net.ParseIP(record).To4()
			answers = append(answers, rr)
		}
	}

	if addr.Is6() {
		for record := range val.AAAA {
			rr := new(dns.AAAA)
			rr.Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeAAAA, Class: dns.ClassINET}
			rr.AAAA = net.ParseIP(record).To6()
			answers = append(answers, rr)
		}
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = answers

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (p Avoid) Name() string { return "avoid" }
