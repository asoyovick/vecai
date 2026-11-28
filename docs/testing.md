# Testing

## Backend

Run all backend checks:

```bash
make check
```

Run tests with coverage:

```bash
make test-cover
```

## Coverage Baseline

Recorded on 2026-11-28 (issue #12):

- `backend/internal/auth`: 33.8%
- `backend/internal/project`: 32.3%
- Other backend packages: 0% (no tests yet)

Integration tests that require PostgreSQL are skipped unless `DATABASE_URL` is set.
