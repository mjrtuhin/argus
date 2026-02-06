# 👁️ ARGUS

**The all-seeing guardian for your infrastructure**

Zero-configuration ML-powered anomaly detection for Prometheus & Grafana.

## 🚧 Status

Under active development. v1.0 launching ~Feb 20, 2026.

**Current Progress:** Week 1, Day 1

## What is Argus?

Argus is the missing ML layer for your observability stack. It automatically:
- 🔍 Discovers all your Prometheus metrics
- 🧠 Learns normal behavior patterns using ML
- 🚨 Detects anomalies in real-time
- 📢 Sends intelligent alerts (no threshold configuration needed)
- 📊 Shows you what broke and why

**Named after the all-seeing giant from Greek mythology with 100 eyes.**

## Features (Coming Soon)

- ✅ Auto-discovers all Prometheus metrics
- ✅ Multi-algorithm ML detection (Prophet + STL + Isolation Forest)
- ✅ Zero manual configuration
- ✅ Adapts to seasonality and trends
- ✅ Smart alerting (Slack, Email, PagerDuty, Webhook)
- ✅ Real-time dashboard
- ✅ One-command Docker deployment
- ✅ Works alongside existing Grafana setup

## Architecture
```
Prometheus → Argus (ML Engine) → Alerts (Slack/Email/PagerDuty)
                ↓
           Dashboard (React)
```

## Tech Stack

- **Backend:** Go (high performance, single binary)
- **ML Engine:** Python (Prophet, scikit-learn)
- **Database:** PostgreSQL
- **Frontend:** React + Recharts
- **Deployment:** Docker

## Quick Start

Coming soon...

## Development

- **Started:** Feb 6, 2026
- **Target Launch:** Feb 20, 2026 (6 weeks)
- **License:** MIT

## Why Argus?

Existing solutions either:
- ❌ Cost £500+/month (Datadog, Dynatrace)
- ❌ Require manual threshold configuration (Grafana alerts)
- ❌ Are too complex to set up (enterprise tools)

Argus is:
- ✅ Free and open-source
- ✅ Automatic (zero config)
- ✅ Simple (one Docker command)

---

**Built with ☕ by MJR Tuhin**

Star ⭐ this repo if you're interested in updates!