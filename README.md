# Cachefik

## Task

We’d like you to write an HTTP reverse proxy, written in Go, that embeds a cache feature. The only requirement is that you don’t use `net/http/httputil`.

What we’ll be looking for in this project is how you structure the code and handle problems, your commit history, and of course the project itself as a whole.

Please share your repository a day before the actual interview so that the team can review your work properly.

During the interview, we'd like you to show us a (working) demo and present the code so that it sparks discussion with the team.

Of course, feel free to add any feature that you find cool & useful, finished or not.

---

## Architecture overview

Cachefik is a small reverse proxy with an embedded HTTP cache.

It:

* receives incoming HTTP requests
* dynamically selects an upstream service
* decides whether a request/response is cacheable
* serves responses either from cache or by proxying upstream
* exposes cache behavior via the `X-Cache` response header

---

## Docker-based service auto-discovery

Instead of hard-coding upstream URLs, Cachefik implements a **simplified Docker provider**, inspired by Traefik’s provider model.

### Discovery model

At startup, Cachefik:

1. Connects to the Docker daemon via the Docker socket
2. Lists running containers
3. Selects containers that opt in via labels
4. Builds an in-memory routing table
5. Applies routing rules based on request path prefixes

Discovery is **startup-only** (no live reload), which keeps the implementation simple and explicit.

### Label contract

Containers opt into discovery using Docker labels:

```text
cachefik.enable=true
cachefik.rule=PathPrefix(`/api`)
cachefik.port=8080
```

Where:

* `cachefik.enable` enables discovery for the container
* `cachefik.rule` defines a routing rule (currently `PathPrefix`)
* `cachefik.port` specifies the container port to route traffic to

### Routing behavior

Routes are matched by **specificity**:

* Longer (more specific) path prefixes are evaluated first
* For example, `/api` takes priority over `/`

This mirrors Traefik’s router priority behavior and avoids order-dependent routing bugs.

### Docker socket access

To enable discovery, the Docker socket is mounted read-only into the proxy container:

```yaml
volumes:
  - /var/run/docker.sock:/var/run/docker.sock:ro
```

This is the same mechanism used by Traefik when running with the Docker provider.

---

## Caching semantics

### A request is cacheable only if **all** of the following are true:

* HTTP method is `GET`
* No `Authorization` header is present
* Request does **not** include `Cache-Control: no-store`

### A response is cacheable only if:

* It does **not** include `Cache-Control: no-store`
* It does **not** include `Cache-Control: private`
* The response status code is cacheable (e.g. `200 OK`)

### TTL handling

* If the response includes `Cache-Control: max-age=N`, that value is used
* Otherwise, a default TTL of **30 seconds** is applied

### Cache behavior indicators

Cachefik adds an `X-Cache` header to responses:

* `X-Cache: HIT`
  The response was served directly from cache. The upstream was not contacted.

* `X-Cache: MISS`
  The request was cacheable, but no cached entry existed. The response was fetched from upstream and stored.

* `X-Cache: BYPASS`
  The request or response was not eligible for caching, so the cache was skipped entirely.

---

## Running the demo (Docker Compose)

The demo runs Cachefik together with two upstream services inside a Docker network:

* **frontend** — cacheable responses
* **backend** — non-cacheable responses (`Cache-Control: no-store`)

### Start the demo

```bash
docker compose up --build
```

Cachefik will be available on:

```text
http://localhost:8000
```

---

## Demo scenarios

### 1. Cacheable frontend (HIT / MISS)

```bash
curl -i http://localhost:8000/
```

* First request → `X-Cache: MISS`
* Subsequent request (within TTL) → `X-Cache: HIT`

The frontend returns a timestamped response, making cache behavior visible.

---

### 2. Non-cacheable backend (BYPASS)

```bash
curl -i http://localhost:8000/api
```

This route always returns:

```text
X-Cache: BYPASS
```

because the backend explicitly sends:

```http
Cache-Control: no-store
```

---

### 3. Authorization header (BYPASS)

```bash
curl -i \
  -H "Authorization: Bearer token" \
  http://localhost:8000/
```

Authenticated requests are bypassed to avoid caching user-specific responses.

---

## Running tests

All cache behavior is covered by unit tests.

```bash
go test ./...
```

Tests verify:

* cache HIT / MISS / BYPASS behavior
* TTL expiration
* request and response cacheability rules

---

## Design choices & tradeoffs

* **No `net/http/httputil`**
  The proxy logic is implemented manually to make request cloning, header handling, and streaming explicit.

* **Docker provider instead of static upstreams**
  Demonstrates a dynamic discovery model similar to Traefik’s provider architecture.

* **In-memory cache only**
  Chosen for simplicity and clarity. No eviction beyond TTL.

* **Conservative caching defaults**
  It is safer to bypass caching than to cache incorrectly.

* **No streaming cache**
  Responses are only buffered when they are cacheable.

---

## Limitations / future work

This project intentionally keeps scope limited. Possible extensions include:

* Live Docker event watching (hot reload)
* LRU or size-bounded cache eviction
* Disk-backed or distributed cache
* Support for `Vary` headers
* Conditional requests (`ETag`, `If-Modified-Since`)
* Structured logging and configurable log sinks
* HTTP/2 upstream support
