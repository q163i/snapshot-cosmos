# Snapshot Cosmos

Multi-node Cosmos blockchain snapshot service with S3 upload support.

## What it does

- Creates compressed snapshots of Cosmos node data
- Uploads to S3 with retention policies
- Supports multiple nodes (Cosmos Hub, Osmosis, Juno, etc.)
- Runs as daemon or one-off commands

## Quick start

```bash
# Build
go build -o snapshot-cosmos .

# List configured nodes
./snapshot-cosmos list

# Create snapshot for specific node
./snapshot-cosmos create cosmoshub

# Run daemon
./snapshot-cosmos daemon cosmoshub
```

## Config

Create `config/nodes.yaml`:

```yaml
nodes:
  cosmoshub:
    enabled: true
    node:
      home_dir: "/home/cosmos/.cosmos"
      chain_id: "cosmoshub-4"
    snapshot:
      interval: "24h"
      retention: 7
      temp_dir: "/tmp/snapshot-cosmos/cosmoshub"
    s3:
      bucket: "q163i-snapshots"
      path_prefix: "snapshots/cosmoshub"
```

## Docker

```bash
# Build
docker build -t snapshot-cosmos .

# Run
docker run -v /home/cosmos/.cosmos:/home/cosmos/.cosmos:ro \
  -e AWS_ACCESS_KEY_ID=your_key \
  -e AWS_SECRET_ACCESS_KEY=your_secret \
  snapshot-cosmos daemon cosmoshub
```

## Kubernetes (Helm)

```bash
# Install with default values
helm install snapshot-cosmos ./helm

# Install with S3 credentials
helm install snapshot-cosmos ./helm \
  --set s3.accessKey=your_key \
  --set s3.secretKey=your_secret

# Enable multiple nodes
helm install snapshot-cosmos ./helm \
  --set config.nodes.cosmoshub.enabled=true \
  --set config.nodes.osmosis.enabled=true
```

## Commands

```bash
snapshot-cosmos list                    # Show configured nodes
snapshot-cosmos create <node>           # Create snapshot
snapshot-cosmos upload <node> <file>    # Upload to S3
snapshot-cosmos daemon <node>           # Run daemon
snapshot-cosmos version                 # Show version
```

## Supported chains

- Cosmos Hub (`cosmoshub`)
- Osmosis (`osmosis`)
- Juno (`juno`)

## S3 structure

```
q163i-snapshots/
└── snapshots/
    ├── cosmoshub/
    │   └── cosmoshub-4-snapshot-2024-01-15-10-30-00.tar.gz
    └── osmosis/
        └── osmosis-1-snapshot-2024-01-15-10-30-00.tar.gz
```

## Environment vars

```bash
SNAPSHOT_COSMOS_NODE_HOME_DIR=/path/to/node
SNAPSHOT_COSMOS_S3_BUCKET=my-bucket
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
```

## Development

```bash
go mod download
go build -o snapshot-cosmos .
go test ./...
```

## License

MIT 