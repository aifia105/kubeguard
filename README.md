# KubeGuard

Kubernetes security auditing and AI-powered troubleshooting tool built with Go.

KubeGuard connects to a Kubernetes cluster to inspect resources, collect logs and events, identify security misconfigurations, and use an LLM to help diagnose issues.

## Features

-  Kubernetes cluster scanning
-  Security auditing
-  Pod logs and event collection
-  AI-powered diagnosis with Ollama + Llama
-  Built with Go and client-go

## Architecture

                    Kubernetes Cluster
                           │
                           ▼
                    ┌─────────────┐
                    │  KubeGuard  │
                    │     Go      │
                    └──────┬──────┘
                           │
             ┌─────────────┼─────────────┐
             │             │             │
             ▼             ▼             ▼
          Scanner        Auditor        Logs
             │             │             │
             └─────────────┼─────────────┘
                           ▼
                       Diagnose
                           │
                           ▼
                    Evidence Layer
                           │
                           ▼
                        Analyze
                           │
                           ▼
                        Ollama
                           │
                           ▼
                         Llama
                           │
                           ▼
                       Diagnosis

## Tech Stack
- Go
- Kubernetes / client-go
- Ollama
- Llama

## Installation

### Prerequisites

- Go 1.21 or later
- Access to a Kubernetes cluster (a valid `kubeconfig`, typically at `~/.kube/config`)
- [Ollama](https://ollama.com) installed and running locally, with a Llama model pulled

### 1. Install Ollama and pull a model

```bash
# Install Ollama (see https://ollama.com/download for OS-specific instructions)
curl -fsSL https://ollama.com/install.sh | sh

# Pull the Llama model KubeGuard uses for diagnosis
ollama pull llama3
```
**Build from source**

```bash
git clone https://github.com/aifia105/kubeguard.git
cd kubeguard
go build -o kubeguard .
sudo mv kubeguard /usr/local/bin/
```

### 3. Verify the installation

```bash
kubeguard --version
kubeguard scan --help
```

### 4. Usage

```bash
Scan command:
kubeguard scan
kubeguard scan pods
kubeguard scan namespaces
...
Audit command:
kubeguard audit
kubeguard audit pods
kubeguard audit deployments
...
Logs command:
kubeguard logs <pod>
...
Diagnose command:
kubeguard diagnose
...
diagnose analyze:
kubeguard analyze --model llama3
```

### Example: scan → audit → diagnose → analyze

A typical end-to-end workflow looks like this:

```bash
# 1. Scan the cluster to build an inventory of resources
kubeguard scan
kubeguard scan pods --namespace production

# 2. Audit those resources for security misconfigurations
kubeguard audit
kubeguard audit deployments --namespace production

# 3. Collect logs/events for a suspicious or failing pod found during the audit
kubeguard logs my-app-6f9c9b7d8-xk2pl --namespace production

# 4. Diagnose the issue using the collected evidence (logs, events, findings)
kubeguard diagnose --namespace production

# 5. Run AI-powered analysis on the diagnosis to get a plain-English explanation
#    and recommended remediation steps
kubeguard analyze --model llama3
```

Each step feeds evidence into the next: `scan` builds the resource inventory, `audit` flags misconfigurations against that inventory, `logs`/`diagnose` gather and correlate runtime evidence, and `analyze` hands the validated evidence to the LLM for reasoning — so the AI is only ever reasoning over facts KubeGuard has already verified.

## Vision

KubeGuard follows an evidence-first approach to Kubernetes analysis:

```text
Kubernetes Cluster
        │
        ▼
   Observation
        │
        ▼
Kubernetes Data
        │
        ├── Logs
        ├── Events
        ├── Resources
        └── Configurations
        │
        ▼
Deterministic Analysis
        │
        ▼
 Security Findings
        │
        ▼
 Relevant Evidence
        │
        ▼
    AI Reasoning
        │
        ▼
    Diagnosis

```

---

KubeGuard collects and validates those facts first, then gives the AI the context required to reason about them.

This makes the system more predictable and reduces unnecessary LLM context.
