# KubeGuard

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Kubernetes client-go](https://img.shields.io/badge/Kubernetes-client--go-326CE5?style=flat&logo=kubernetes)](https://github.com/kubernetes/client-go)
[![AI Engine](https://img.shields.io/badge/AI-Ollama-black?style=flat&logo=ollama)](https://ollama.com)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

**KubeGuard** is a Kubernetes security auditing, resource inventory, and AI-powered troubleshooting tool built in Go. It connects to your Kubernetes cluster to inspect resources, detect security misconfigurations, gather runtime evidence, and leverage local LLMs via Ollama to provide root-cause analysis and actionable remediations.

---

##  Key Features

- **Concurrent Resource Scanning (`scan`)**: Build a complete cluster inventory across 15+ resource types (Pods, Nodes, Namespaces, Deployments, Services, ConfigMaps, Ingresses, Secrets, LimitRanges, ResourceQuotas, NetworkPolicies, and Node/Pod Metrics). Secret values are automatically redacted in JSON exports.
- **Deterministic Security Auditing (`audit`)**: Run 20+ automated security rules across 8 resource categories, highlighting vulnerabilities across `CRITICAL`, `HIGH`, `MEDIUM`, and `LOW` severity levels.
- **Log & Event Aggregation (`logs`)**: Fetch, tail, follow, and search container logs with cross-namespace pod auto-discovery.
- **Evidence Snapshot Engine (`diagnose`)**: Correlate unhealthy pod statuses, container exit codes, cluster warning events, pod logs, and security audit findings into a structured evidence snapshot (`output/diagnose_results.json`).
- **Evidence-First AI Diagnosis (`analyze`)**: Feed gathered evidence snapshots directly to Ollama LLMs (e.g. `phi3`, `llama3`) for plain-English explanations and step-by-step resolution advice without cluttering context windows.
- **PostgreSQL History & Persistence (`db`)**: Store cluster snapshots, audit runs, findings, and LLM diagnoses over time with automated database schema migration support (`golang-migrate`).
- **OpenSearch Indexing (`search`)**: Create indices (`findings`, `diagnostics`, `events`) and sync findings, diagnostics, and events from PostgreSQL into OpenSearch for fast indexing and search analytics.

---

## Architecture & Philosophy

KubeGuard follows an **Evidence-First** approach to Kubernetes diagnostic AI:

```text
               +----------------------------------+
               |        Kubernetes Cluster        |
               +----------------------------------+
                                |
               +----------------+----------------+
               |                                 |
               v                                 v
     +-------------------+             +-------------------+
     |   kubeguard scan  |             |  kubeguard audit  |
     | (Inventory/Metrics)             | (Security Checks) |
     +---------+---------+             +---------+---------+
               |                                 |
               +----------------+----------------+
                                |
                                v
                   +------------------------+
                   |   kubeguard diagnose   |
                   | (Gather Evidence: Logs,|
                   |  Events, Findings)     |
                   +-----------+------------+
                                |
                                v
                    output/diagnose_results.json
                                |
                                v
                   +------------------------+
                   |   kubeguard analyze    |
                   |    (Ollama / Llama3)   |
                   +-----------+------------+
                                |
                                v
               Root-Cause Diagnosis & Remediation
```

Rather than feeding raw, unorganized cluster dumps to an LLM, KubeGuard performs **deterministic facts-first collection**:
1. **Gather**: Inspect K8s objects, events, metrics, and logs deterministically.
2. **Audit**: Match resources against defined security rules.
3. **Filter**: Retain only failing resources, warning events, and unhealthy logs.
4. **Reason**: Pass the validated, noise-free evidence bundle to the LLM.

---

## Tech Stack

- **Language**: Go 1.21+
- **Kubernetes Client**: `client-go` and `metrics` client
- **CLI Framework**: `spf13/cobra`
- **Database**: PostgreSQL (via `pgx/v5`) and schema migrations (`golang-migrate`)
- **Search Engine**: OpenSearch (via `opensearch-go`)
- **AI Integration**: Ollama REST API (Llama3, Phi3, etc.)

---

## Installation & Setup

### Prerequisites

1. **Go**: Version 1.21 or later installed.
2. **Kubernetes Access**: A valid `kubeconfig` (typically located at `~/.kube/config`).
3. **Ollama** *(Optional for AI analysis)*: Installed and running locally.

### 1. Build from Source

```bash
# Clone the repository
git clone https://github.com/aifia105/kubeguard.git
cd kubeguard

# Build the binary
go build -o kubeguard .

# (Optional) Move to system PATH
sudo mv kubeguard /usr/local/bin/
```

### 2. Set Up Ollama (For AI Analysis)

```bash
# Install Ollama (Linux/macOS)
curl -fsSL https://ollama.com/install.sh | sh

# Pull your preferred model (e.g. phi3 or llama3)
ollama pull phi3
```

### 3. Setup Database (Optional)

KubeGuard can automatically persist scan results, audit findings, and diagnoses to PostgreSQL.

Set your database connection URL in `.env` or environment variables:
```bash
export KUBEGUARD_POSTGRES_URL="postgres://user:password@localhost:5432/kubeguard?sslmode=disable"
```

Run database migrations:
```bash
kubeguard db migrate-up
```

### 4. Setup OpenSearch (Optional)

KubeGuard can index findings, diagnostics, and cluster events into OpenSearch.

Set your OpenSearch environment variables:
```bash
export KUBEGUARD_OPENSEARCH_URL="http://localhost:9200"
export KUBEGUARD_OPENSEARCH_USER="admin"
export KUBEGUARD_OPENSEARCH_PASSWORD="admin"
```

Initialize the OpenSearch indices (`findings`, `diagnostics`, `events`):
```bash
kubeguard search setup
```

---

## Command Reference

### Global Flags

| Flag          | Short | Description                             | Default        |
| ------------- | ----- | --------------------------------------- | -------------- |
| `--namespace` | `-n`  | Scope operation to a specific namespace | All namespaces |
| `--output`    | `-o`  | Output format (`text` or `json`)        | `text`         |
| `--no-save`   |       | Skip persisting results to the database | `false`        |

---

### `kubeguard scan`

Concurrently scans the cluster and displays resource inventories and metrics.

```bash
# Scan full cluster
kubeguard scan

# Scan specific resources
kubeguard scan pods -n production
kubeguard scan nodes
kubeguard scan deployments
kubeguard scan services
kubeguard scan ingresses
kubeguard scan secrets
kubeguard scan configmaps
kubeguard scan events
kubeguard scan metrics
kubeguard scan networkpolicies
kubeguard scan limitranges
kubeguard scan resourcequotas

# Output JSON (redacts secret contents)
kubeguard scan -o json
```

---

### `kubeguard audit`

Audits cluster resources against security best practices and misconfiguration checks.

```bash
# Run full cluster audit
kubeguard audit

# Audit specific resource types
kubeguard audit pods -n production
kubeguard audit deployments
kubeguard audit services
kubeguard audit ingresses
kubeguard audit secrets
kubeguard audit configmaps
kubeguard audit namespaces

# Export findings to JSON
kubeguard audit -o json
```

#### Security Audit Checks Included:

| Resource       | Rule ID                                       | Severity   | Description                                                               |
| -------------- | --------------------------------------------- | ---------- | ------------------------------------------------------------------------- |
| **Pod**        | `PRIVILEGED_CONTAINER`                        | `HIGH`     | Containers running with privileged access                                 |
| **Pod**        | `RUNNING_AS_ROOT`                             | `HIGH`     | Containers running as root user (UID 0 or non-root false)                 |
| **Pod**        | `PRIVILEGE_ESCALATION_ALLOWED`                | `HIGH`     | Containers allowing privilege escalation                                  |
| **Pod**        | `WRITABLE_ROOT_FILESYSTEM`                    | `HIGH`     | Root filesystem is writable                                               |
| **Pod**        | `DANGEROUS_CAPABILITIES_KEPT`                 | `HIGH`     | Retaining capabilities like `NET_ADMIN`, `SYS_ADMIN`, etc.                |
| **Pod**        | `HOST_NAMESPACES_SHARED`                      | `HIGH`     | Pod sharing HostNetwork, HostPID, or HostIPC                              |
| **Pod**        | `SENSITIVE_HOST_PATH_MOUNTED`                 | `HIGH`     | Mounting sensitive host paths (`/var/run/docker.sock`, `/etc`, `/`, etc.) |
| **Pod**        | `ENV_VARS_WITH_PLAINTEXT_SECRETS`             | `HIGH`     | Sensitive variable names defined in plaintext                             |
| **Pod**        | `DEFAULT_SA_AUTO_MOUNTED`                     | `MEDIUM`   | Automounted service account token on default SA                           |
| **Pod**        | `NO_RESOURCE_LIMITS` / `REQUESTS`             | `MEDIUM`   | Missing resource limits or requests                                       |
| **Pod**        | `LATEST_IMAGE_TAG`                            | `MEDIUM`   | Container image using `:latest` or missing tag                            |
| **Pod**        | `NO_SECCOMP_PROFILE` / `NO_APP_ARMOR_PROFILE` | `MEDIUM`   | Missing security profiles                                                 |
| **Pod**        | `NO_LIVENESS_PROBE` / `NO_READINESS_PROBE`    | `LOW`      | Missing health probes                                                     |
| **Secret**     | `OLD_OR_STALE_SECRET`                         | `HIGH`     | Secrets created > 30 days ago                                             |
| **Secret**     | `UNREFERENCED_SECRET`                         | `MEDIUM`   | Secrets not referenced by any Pod                                         |
| **Secret**     | `LEGACY_SERVICE_ACCOUNT_TOKEN`                | `MEDIUM`   | Legacy ServiceAccountToken secret types                                   |
| **ConfigMap**  | `SECRET_LOOKING_DATA`                         | `HIGH`     | ConfigMaps containing sensitive keys/passwords                            |
| **ConfigMap**  | `UNREFERENCED_CONFIGMAP`                      | `MEDIUM`   | ConfigMaps not referenced by any Pod                                      |
| **Deployment** | `SINGLE_REPLICA_NO_REDUNDANCY`                | `MEDIUM`   | Deployment running only 1 replica                                         |
| **Deployment** | `ROLLING_UPDATE_STRATEGY`                     | `LOW`      | Non-rolling update deployment strategy                                    |
| **Ingress**    | `NO_TLS_CONFIGURED`                           | `HIGH`     | Ingress missing TLS configuration                                         |
| **Ingress**    | `WILDCARD_HOST_CONFIGURED`                    | `MEDIUM`   | Ingress using wildcard hosts                                              |
| **Namespace**  | `NO_DEFAULT_DENY_NETWORK_POLICY`              | `HIGH`     | Namespace missing default deny network policy                             |
| **Namespace**  | `NO_RESOURCE_QUOTAS`                          | `MEDIUM`   | Namespace missing ResourceQuota                                           |
| **Namespace**  | `NO_LIMIT_RANGES`                             | `MEDIUM`   | Namespace missing LimitRange                                              |
| **Node**       | `NODE_NOT_READY`                              | `CRITICAL` | Node in `NotReady` status                                                 |
| **Node**       | `DISK_MEMORY_PID_PRESSURE`                    | `HIGH`     | Node reporting disk, memory, or PID pressure                              |
| **Node**       | `OUTDATED_KUBELET_VERSION`                    | `LOW`      | Outdated Kubelet version                                                  |

---

### `kubeguard logs`

Retrieves container logs with cross-namespace search capabilities.

```bash
# Fetch last 10 lines of a pod (auto-discovers namespace)
kubeguard logs my-pod-name

# Specify namespace, container, tail lines, and follow
kubeguard logs my-pod-name production -c app-container -t 50 -f
```

---

### `kubeguard diagnose`

Gather evidence (unhealthy pod status, logs, events, audit findings) into an evidence bundle (`output/diagnose_results.json`).

```bash
# Generate evidence bundle for the entire cluster
kubeguard diagnose

# Scope evidence bundle to a namespace
kubeguard diagnose -n staging
```

---

### `kubeguard analyze`

Sends the gathered evidence snapshot to an Ollama LLM model for AI-driven root cause diagnosis.

```bash
# Analyze using default model (phi3)
kubeguard analyze

# Use a custom model (e.g. llama3)
kubeguard analyze --model llama3

# Specify custom evidence JSON path and save to DB
kubeguard analyze --model llama3 --file output/diagnose_results.json --save
```

---

### `kubeguard db`

Manage database schema migrations and status.

```bash
# Check current migration status
kubeguard db status

# Apply pending migrations
kubeguard db migrate-up

# Rollback recent migrations
kubeguard db migrate-down --steps 1
```

---

### `kubeguard search`

Manage and sync OpenSearch indices for cluster findings, diagnostics, and events.

```bash
# Create OpenSearch indices (findings, diagnostics, events) if they don't exist
kubeguard search setup

# Sync findings, diagnostics, and events from PostgreSQL into OpenSearch
kubeguard search index

# Query findings in OpenSearch by term, namespace, and result limit
kubeguard search query "privileged" --index findings --namespace production --limit 20

```

---

##  Step-by-Step Workflow Example

A typical end-to-end audit & AI troubleshooting workflow:

```bash
# 1. Scan your cluster namespace to inspect resources
kubeguard scan -n production

# 2. Audit resources for security misconfigurations and vulnerabilities
kubeguard audit -n production

# 3. Check logs for a failing or suspicious pod
kubeguard logs payment-service-7f8d9b4c-2x9z1 -n production -t 100

# 4. Gather a correlated evidence snapshot (findings + unhealthy pod logs + events)
kubeguard diagnose -n production

# 5. Ask the AI engine to analyze the evidence snapshot and recommend fix steps
kubeguard analyze --model phi3 --save

# 6. Set up OpenSearch indices and sync database records to OpenSearch
kubeguard search setup
kubeguard search index
kubeguard search query "privileged" --index findings --namespace production --limit 20
```

---

## Screenshots 
<img width="1882" height="728" alt="Screenshot 2026-09-01 092211" src="https://github.com/user-attachments/assets/7ff38740-5bc7-4b7c-8ca2-21a82334c700" />
<img width="1918" height="562" alt="Screenshot 2026-09-01 092106" src="https://github.com/user-attachments/assets/f71a6560-0084-4684-9a08-aa017888d67a" />

