package eventsource

import (
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-kad-dht/providers"
)

var metrics = "/metrics"

// root namespace of provider keys
var providersRoot = datastore.NewKey(providers.ProvidersKeyPrefix)
var agentsRoot = datastore.NewKey("/agents")
var skillsRoot = datastore.NewKey("/skills")
var metricsRoot = datastore.NewKey("/metrics")

func isProviderKey(k datastore.Key) bool {
	// not interested if this is not a query for providers of a particular cid
	// we're looking for /providers/cid, not /providers (currently used in GC)
	//if !providersRoot.IsAncestorOf(k) || len(k.Namespaces()) < 2 {
	//	return false
	//}

	// ignore keys that are not in the providers namespace
	if !metricsRoot.IsAncestorOf(k) {
		return false
	}

	return true
}
