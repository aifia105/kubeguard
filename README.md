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
Go
Kubernetes / client-go
Ollama
Llama

## Usage
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
diagnose command:
kubeguard diagnose
diagnose analyze:
kubeguard analyze
```

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