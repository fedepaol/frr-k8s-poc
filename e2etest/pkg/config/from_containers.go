package config

import (
	frrk8sv1alpha1 "github.com/metallb/frrk8s/api/v1alpha1"
	"go.universe.tf/e2etest/pkg/frr"
	frrcontainer "go.universe.tf/e2etest/pkg/frr/container"
	"go.universe.tf/e2etest/pkg/ipfamily"
)

func NeighborsForContainers(frrs []*frrcontainer.FRR, modify ...func(frr.Neighbor)) []frrk8sv1alpha1.Neighbor {
	res := make([]frrk8sv1alpha1.Neighbor, 0)
	for _, f := range frrs {
		addresses := f.AddressesForFamily(ipfamily.IPv4)
		ebgpMultihop := false
		if f.NeighborConfig.MultiHop && f.NeighborConfig.ASN != f.RouterConfig.ASN {
			ebgpMultihop = true
		}

		for _, address := range addresses {
			neigh := frrk8sv1alpha1.Neighbor{
				ASN:          f.RouterConfig.ASN,
				Address:      address,
				Port:         f.RouterConfig.BGPPort,
				EBGPMultiHop: ebgpMultihop,
			}
			res = append(res, neigh)
		}
	}
	return res
}
