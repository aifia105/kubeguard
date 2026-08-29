# KubeGuard
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Kubernetes client-go](https://img.shields.io/badge/Kubernetes-client--go-326CE5?style=flat&logo=kubernetes)](https://github.com/kubernetes/client-go)
[![AI Engine](https://img.shields.io/badge/AI-Ollama-black?style=flat&logo=ollama)](https://ollama.com)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
**KubeGuard** is a Kubernetes security auditing, resource inventory, and AI-powered troubleshooting tool built in Go. It connects to your Kubernetes cluster to inspect resources, detect security misconfigurations, gather runtime evidence, and leverage local LLMs via Ollama to provide root-cause analysis and actionable remediations.
---
## 🌟 Key Features
- **🔍 Concurrent Resource Scanning (`scan`)**: Build a complete cluster inventory across 15+ resource types (Pods, Nodes, Namespaces, Deployments, Services, ConfigMaps, Ingresses, Secrets, LimitRanges, ResourceQuotas, NetworkPolicies, and Node/Pod Metrics). Secret values are automatically redacted in JSON exports.
- **🛡️ Deterministic Security Auditing (`audit`)**: Run 20+ automated security rules across 8 resource categories, highlighting vulnerabilities across `CRITICAL`, `HIGH`, `MEDIUM`, and `LOW` severity levels.
- **📜 Log & Event Aggregation (`logs`)**: Fetch, tail, follow, and search container logs with cross-namespace pod auto-discovery.
- **🧾 Evidence Snapshot Engine (`diagnose`)**: Correlate unhealthy pod statuses, container exit codes, cluster warning events, pod logs, and security audit findings into a structured evidence snapshot (`output/diagnose_results.json`).
- **🤖 Evidence-First AI Diagnosis (`analyze`)**: Feed gathered evidence snapshots directly to Ollama LLMs (e.g. `phi3`, `llama3`) for plain-English explanations and step-by-step resolution advice without cluttering context windows.
- **💾 PostgreSQL History & Persistence (`db`)**: Store cluster snapshots, audit runs, findings, and LLM diagnoses over time with automated database schema migration support (`golang-migrate`).
---
## 🏗️ Architecture & Philosophy
KubeGuard follows an **Evidence-First** approach to Kubernetes diagnostic AI:
```text
               +----------------------------------+
               |        Kubernetes Cluster        |
