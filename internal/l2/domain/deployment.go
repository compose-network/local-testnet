package domain

// DeploymentState represents the L1 deployment state (from state.json)
// NOTE: This is a partial representation - only fields actually used by the application are included.
// The full state.json file contains additional fields that are omitted here.
type DeploymentState struct {
	ImplementationsDeployment ImplementationsDeployment `json:"implementationsDeployment"`
	OpChainDeployments        []OpChainDeployment       `json:"opChainDeployments"`
}

// ImplementationsDeployment represents shared implementation contracts
type ImplementationsDeployment struct {
	DisputeGameFactoryImplAddress string `json:"disputeGameFactoryImplAddress"`
}

// OpChainDeployment represents per-chain L1 contracts
type OpChainDeployment struct {
	ID                             string     `json:"id"`
	SystemConfigProxyAddress       string     `json:"systemConfigProxyAddress"`
	L1StandardBridgeProxyAddress   string     `json:"l1StandardBridgeProxyAddress"`
	OptimismPortalProxyAddress     string     `json:"optimismPortalProxyAddress"`
	DisputeGameFactoryProxyAddress string     `json:"disputeGameFactoryProxyAddress"`
	StartBlock                     StartBlock `json:"startBlock"`
}

// StartBlock represents the L1 block 
// where the rollup starts
type StartBlock struct {
	Hash   string `json:"hash"`
	Number string `json:"number"`
}
