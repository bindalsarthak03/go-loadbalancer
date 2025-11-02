# ⚖️ Go Load Balancer

A simple yet powerful **HTTP Load Balancer** written in Go — built as a learning project to understand **reverse proxies**, **health checks**, and **go routines**.

---

## 🚀 Features

✅ **Round Robin Load Balancing** — distributes incoming traffic evenly across multiple backend servers.
✅ **Health Checks** — periodically checks backend `/health` endpoints and updates status automatically.
✅ **Metrics & Dashboard** — exposes `/status` endpoint with backend health and request count data.
✅ **Concurrent Safe** — uses `sync/atomic` for efficient counter and state management.

---

## 🏗️ Project Structure

```
go-loadbalancer/
├── loadbalancer/
│   ├── main.go               # Main load balancer entry point
│   ├── healthchecker.go      # Backend health monitoring
│   ├── config.yaml           # YAML configuration for backends
│   ├── config.go             # Reads YAML and loads config
│   └── metrics.go            # Tracks requests and status
└── README.md
```
---

## 🧠 How It Works

1. **Health Checker** — periodically pings each backend’s `/health` endpoint.
2. **Round Robin Selector** — selects the next healthy backend in order using atomic counters.
3. **Reverse Proxy** — forwards incoming requests via `httputil.NewSingleHostReverseProxy`.
4. **Metrics Collector** — tracks number of requests handled by each backend.
5. **Status Dashboard** — exposes `/status` endpoint with backend health in JSON format.

---

## 💻 Run Locally

### 1️⃣ Start 3 Simple Backends

You can spin up dummy servers for testing

### 2️⃣ Start Load Balancer

```bash
cd loadbalancer
go run main.go
```

### 3️⃣ Test the Balancer

Visit:

* Load Balancer → `http://localhost:8080`
* Dashboard → `http://localhost:8080/status`

---

## 📊 Example `/status` Output

```json
{
  "backends": [
    {"url": "http://localhost:8081", "healthy": true, "requests": 12},
    {"url": "http://localhost:8082", "healthy": false, "requests": 0},
    {"url": "http://localhost:8083", "healthy": true, "requests": 9}
  ]
}
```
---

## 🧰 Tech Stack

* **Language:** Go (v1.21+)
* **Libraries:**

  * `net/http`
  * `net/http/httputil`
  * `gopkg.in/yaml.v3`
  * `sync/atomic`

---

## 👨‍💻 Author

**Sarthak Bindal**\
Built to understand the internals of load balancing and system design.

