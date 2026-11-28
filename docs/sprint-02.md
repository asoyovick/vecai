# Sprint 02 – Stabilization and Core MVP Flows

## Sprint Goal

Harden the MVP with security fixes, automated tests, CI, and complete the first end-to-end frontend project management flows.

## Duration

2 weeks (planned)

## Planned Scope

### P0 Security and Quality

- [x] Add backend tests for auth and project access (#12)
- [ ] Replace mock email verification with a real flow (#13)
- [x] Fix project list visibility for member projects (#14)
- [ ] Move document uploads from local disk to object storage (#15)

### P1 Frontend and CI

- [x] Project detail page with tabs (#16)
- [x] Materials tracking UI (#17)
- [x] Daily site reports UI (#18)
- [x] CI workflow for backend checks and frontend lint/build (#20)
- [ ] BOQ management and comparison UI (#23)
- [ ] Documents upload and download UI (#24)

### P2 Supporting Work

- [x] Add update endpoint for daily reports (#21)
- [x] Record Sprint 01 outcome and plan Sprint 02 (#25)
- [ ] Dockerfile, deployment config and backups (#26)

## Definition of Done

- `make check` passes in CI.
- Core project detail tabs have usable CRUD flows against real API endpoints.
- Auth and project access rules are covered by automated tests.
- Deployment/documentation reflects actual runtime behavior.

## Notes

Sprint 02 planning is based on open GitHub issues and the current MVP roadmap. Items completed during stabilization are checked off above.
