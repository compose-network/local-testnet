Clone required repos into the same parent directory:

```
~/your-workspace/
├── compose-publisher/   branch: feature/compose-sidecar
├── op-geth/             tag:    v1.101603.4
└── sidecar/             branch: stage
```

```bash
# compose-publisher
git clone git@github.com:compose-network/compose-publisher.git
cd compose-publisher && git checkout feature/compose-sidecar

# op-geth
git clone git@github.com:compose-network/op-geth.git
cd op-geth && git checkout v1.101603.4

# sidecar
git clone git@github.com:compose-network/sidecar.git
cd sidecar && git checkout stage

# op-rbuilder (optional local override; default is remote compose-network/op-rbuilder#stage)
# git clone git@github.com:compose-network/op-rbuilder.git
# cd op-rbuilder && git checkout stage
```

## Run

```bash
cd local-testnet
# default op-rbuilder source: https://github.com/compose-network/op-rbuilder.git#stage
make run-l2 L2_ARGS="--flashblocks-enabled --sidecar-enabled --blockscout-enabled"

# optional local override
# OP_RBUILDER_PATH=../../op-rbuilder make run-l2 L2_ARGS="--flashblocks-enabled --sidecar-enabled --blockscout-enabled"
```
