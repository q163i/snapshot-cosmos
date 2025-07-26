# Snapshot Cosmos Helm Chart

Helm chart for deploying Snapshot Cosmos service to Kubernetes.

## Quick start

```bash
# Install with default values
helm install snapshot-cosmos ./helm

# Install with custom values
helm install snapshot-cosmos ./helm \
  --set s3.accessKey=your_key \
  --set s3.secretKey=your_secret \
  --set config.nodes.cosmoshub.enabled=true
```

## Configuration

### S3 credentials

```bash
helm install snapshot-cosmos ./helm \
  --set s3.accessKey=AKIA... \
  --set s3.secretKey=your_secret_key
```

### Enable multiple nodes

```bash
helm install snapshot-cosmos ./helm \
  --set config.nodes.cosmoshub.enabled=true \
  --set config.nodes.osmosis.enabled=true \
  --set volumes.osmosis.enabled=true
```

### Custom image

```bash
helm install snapshot-cosmos ./helm \
  --set image.repository=my-registry/snapshot-cosmos \
  --set image.tag=v1.0.0
```

## Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `snapshot-cosmos` |
| `image.tag` | Image tag | `latest` |
| `s3.accessKey` | S3 access key | `""` |
| `s3.secretKey` | S3 secret key | `""` |
| `config.nodes.cosmoshub.enabled` | Enable Cosmos Hub | `true` |
| `config.nodes.osmosis.enabled` | Enable Osmosis | `false` |
| `volumes.cosmoshub.enabled` | Mount Cosmos Hub data | `true` |
| `volumes.osmosis.enabled` | Mount Osmosis data | `false` |
| `daemon.enabled` | Run in daemon mode | `true` |
| `daemon.node` | Node to snapshot | `cosmoshub` |

## Storage

The chart creates PVCs for node data:

- `cosmos-data` - Cosmos Hub data
- `osmosis-data` - Osmosis data (if enabled)

## Security

- Runs as non-root user (UID 1001)
- Read-only root filesystem
- Dropped capabilities
- ServiceAccount for RBAC

## Monitoring

- Liveness probe on port 8080
- Readiness probe on port 8080
- Resource limits and requests 