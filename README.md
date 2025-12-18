# Cachefik

## Cache
Cache only when all are true:
- Method is GET (optionally HEAD)
- No Authorization header
- No Cache-Control: no-store
- Response status is 200

TTL:
- Use Cache-Control: max-age
- Fallback to a default (e.g. 30s)

## WIP
- Add a configurable logpath for the proxy, instead of logging everything to stdout
- Modify cache to use LRU