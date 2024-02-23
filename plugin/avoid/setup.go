/*
 * This code comes from the coredns repo: https://github.com/coredns/demo
 * under apache 2 license.
 */
package avoid

import (
	"strconv"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	log "github.com/sirupsen/logrus"
)

var (
	pName = "avoid"
)

func init() {
	caddy.RegisterPlugin("avoid", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	c.Next() // 'avoid'

	args := c.RemainingArgs()

	// avoid dns-service hostname, portname
	if len(args) > 2 {
		return plugin.Error(pName, c.ArgErr())
	}

	// working backwards if we have 2 values, the second is the port num
	if len(args) > 1 {
		portNum, err := strconv.ParseUint(args[1], 10, 16)
		if err != nil {
			return plugin.Error(pName, err)
		}
		AvoidDNSServerPort = int(portNum)

		log.Infof("%s: Set avoid backend port to: %d\n", pName, AvoidDNSServerPort)
	}

	// if we only have 1 value, then it is our hostname
	if len(args) > 0 {
		// TODO: Some sanity checks here?
		AvoidDNSServerHost = args[0]

		log.Infof("%s: Set avoid backend hostname to: %s\n", pName, AvoidDNSServerHost)
	}

	if c.NextArg() {
		return plugin.Error(pName, c.ArgErr())
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return Avoid{}
	})

	return nil
}
