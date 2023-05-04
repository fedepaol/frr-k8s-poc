package frr

import (
	"fmt"
	"net"

	v1alpha1 "github.com/metallb/frrk8s/api/v1alpha1"
	"github.com/metallb/frrk8s/internal/ipfamily"
)

func routerToFRRConfig(r v1alpha1.Router) (*routerConfig, error) {
	res := &routerConfig{
		MyASN:        r.ASN,
		RouterID:     r.ID,
		VRF:          r.VRF,
		Neighbors:    make([]*neighborConfig, 0),
		IPV4Prefixes: make([]string, 0),
		IPV6Prefixes: make([]string, 0),
	}

	for _, n := range r.Neighbors {
		frrNeigh, err := neighborToFRR(n)
		if err != nil {
			return nil, err
		}
		res.Neighbors = append(res.Neighbors, frrNeigh)
	}
	for _, p := range r.PrefixesV4 {
		res.IPV4Prefixes = append(res.IPV4Prefixes, p)
	}
	for _, p := range r.PrefixesV6 {
		res.IPV6Prefixes = append(res.IPV6Prefixes, p)
	}
	return res, nil
}

func neighborToFRR(n v1alpha1.Neighbor) (*neighborConfig, error) {
	neighborFamily, err := ipfamily.ForAddresses(n.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to find ipfamily for %s, %w", n.Address, err)
	}
	res := &neighborConfig{
		Name:           neighborName(n.ASN, n.Address),
		ASN:            n.ASN,
		Addr:           n.Address,
		Port:           n.Port,
		Password:       n.Password,
		Advertisements: make([]*advertisementConfig, 0),
		IPFamily:       neighborFamily,
	}
	// TODO allow all
	for _, p := range n.AllowedOutPrefixes.Prefixes {
		_, cidr, err := net.ParseCIDR(p)
		if err != nil {
			return nil, err
		}
		family := ipfamily.ForCIDR(cidr)

		switch family {
		case ipfamily.IPv4:
			res.HasV4Advertisements = true
		case ipfamily.IPv6:
			res.HasV6Advertisements = true
		}
		res.Advertisements = append(res.Advertisements, &advertisementConfig{Prefix: p, IPFamily: family})
	}

	return res, nil
}

func neighborName(ASN uint32, peerAddr string) string {
	return fmt.Sprintf("%d@%s", ASN, peerAddr)
}
