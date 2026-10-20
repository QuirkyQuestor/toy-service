# Toy Service

## Run

```
docker compose up -d
docker compose exec -T db psql -U postgres -d toy_service < migration/001-ItemsTable.sql
export DATABASE_URL=postgres://postgres:password@localhost:5432/toy_service
make run
```

## Test
```
curl -v http://localhost:8080/healthz
curl -v http://localhost:8080/readyz
```

## Status
Working: health and readiness probes, graceful shutdown, GET /items/{id}.<br>
Next: remaining CRUD handlers, migrations, tests.

[Plan](plan.md)