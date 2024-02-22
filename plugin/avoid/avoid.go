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

	"github.com/coredns/coredns/request"
	pkg "github.com/isi-lincoln/avoid/pkg"
	avoid "github.com/isi-lincoln/avoid/protocol"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

var (
	avoidDNSServerHost = "avoid"
	avoidDNSServerPort = pkg.DefaultAvoidDNSPort
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

	ue := state.IP()

	var entry *avoid.DNSEntry
	pkg.WithAvoidDNS(
		fmt.Sprintf("%s:%d", avoidDNSServerHost, avoidDNSServerPort),
		func(c avoid.DNSClient) error {

			log.Infof("requesting: %s/%s from %s:%d", ue, qname, avoidDNSServerHost, avoidDNSServerPort)

			resp, err := c.Show(context.TODO(), &avoid.ShowRequest{
				Ue:   ue,
				Name: qname,
			})

			if err != nil {
				log.Error(err)
			}

			entry = resp.Entry

			return nil
		})

	answers := []dns.RR{}
	for _, v4record := range entry.Arecords {
		rr := new(dns.A)
		rr.Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeA, Class: dns.ClassINET}
		rr.A = net.ParseIP(v4record).To4()
		answers = append(answers, rr)
	}

	for _, v6record := range entry.Aaaarecords {
		rr := new(dns.AAAA)
		rr.Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeAAAA, Class: dns.ClassINET}
		rr.AAAA = net.ParseIP(v6record).To16()
		answers = append(answers, rr)
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
