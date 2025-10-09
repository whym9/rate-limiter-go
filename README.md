# 🚦 Rate Limiter API (Go + Redis)

A lightweight, production-style **rate limiting service** written in Go.  
Implements a **Fixed Window Counter** algorithm with Redis for distributed state, exposes HTTP endpoints, Prometheus metrics, and Dockerized deployment.

---

## 🧩 Features

- ⚙️ **Fixed Window rate limiter** using Redis `INCR` + `EXPIRE`
- 📡 HTTP API built with **Echo**
- 📊 **Prometheus metrics** (`/metrics`)
- 🧠 Graceful shutdown & structured logging
- 🧪 **Unit + integration tests** (with `miniredis` or `testcontainers-go`)
- 🐳 **Docker Compose** stack (API + Redis)
- 🔄 **CI/CD** pipeline with GitHub Actions
- 🩺 `/healthz` endpoint for monitoring

---

## 📁 Project Structure

![CI](https://github.com/<your-username>/rate-limiter-go/actions/workflows/ci.yml/badge.svg)
