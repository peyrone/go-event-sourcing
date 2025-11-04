# Go Event Sourcing

A simple hands-on project to learn **Event Sourcing** and **CQRS** using **Golang**.

---

## Features
- Full event sourcing loop (commands → events → projections → queries)
- Local SQLite database
- Simple HTTP API for testing
- Clear folder structure for learning

---

## Install GVM (Go Version Manager)

```bash
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
source ~/.gvm/scripts/gvm
```

Verify installation:
```bash
gvm version
```

---

## Install Go Versions

Go 1.24 requires at least Go 1.22.6 to build.  
Install both and set up defaults:

```bash
gvm install go1.22.6
gvm use go1.22.6
export GOROOT_BOOTSTRAP=$GOROOT

gvm install go1.24.2
gvm use go1.24.2 --default
go version
```

You should see:
```
go version go1.24.2 darwin/arm64
```

---

## Clone and Setup

```bash
git clone https://github.com/peyrone/go-event-sourcing.git
cd go-event-sourcing
go mod tidy
```

Run the server:
```bash
go run ./cmd/api
```

Expected output:
```
listening on :8080
```

---

## Try It Out

```bash
# Create
curl -s -XPOST localhost:8080/orders/create \
 -d '{"ID":"o-1001","Customer":"Alice"}' -H 'Content-Type: application/json'

# Add items
curl -s -XPOST localhost:8080/orders/add-item \
 -d '{"OrderID":"o-1001","SKU":"book-1","Qty":2}' -H 'Content-Type: application/json'
curl -s -XPOST localhost:8080/orders/add-item \
 -d '{"OrderID":"o-1001","SKU":"pen-1","Qty":1}' -H 'Content-Type: application/json'

# Remove one
curl -s -XPOST localhost:8080/orders/remove-item \
 -d '{"OrderID":"o-1001","SKU":"book-1","Qty":1}' -H 'Content-Type: application/json'

# Query summary (projection)
curl -s "localhost:8080/orders/summary?id=o-1001" | jq
# => {"orderId":"o-1001","customer":"Alice","totalItems":2,"checkedOut":false}

# Checkout
curl -s -XPOST localhost:8080/orders/checkout \
 -d '{"ID":"o-1001"}' -H 'Content-Type: application/json'

# Query again
curl -s "localhost:8080/orders/summary?id=o-1001" | jq
# => checkedOut: true

---

## Inspect Stored Events

sqlite3 events.db 'SELECT aggregate_id, version, name, datetime(occurred_at, "unixepoch") FROM events ORDER BY id;'
```

---

## Project Structure

```
cmd/api/            → HTTP server entry point
internal/domain/    → Aggregates & events
internal/store/     → SQLite event store
internal/readmodel/ → Projection (read model)
internal/util/      → Utility (id.go for unique IDs)
```

---

## Example Flow

```
Client → Command API → Event Store → Projector → Read Model → Query API
```

