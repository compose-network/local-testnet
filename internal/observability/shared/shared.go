package shared

import "github.com/docker/docker/api/types/container"

const (
	ObservabilityNetworkName = "observability-net"
	L1NetworkName            = "kt-localnet"
	L2NetworkName            = "localnet-l2"
)

var (
	NetworkMode = container.NetworkMode(ObservabilityNetworkName)
	Labels      = map[string]string{"stack": "localnet-observability"}
)
