# Product requirements: Office foosball stats

## Summary

A small internal web app for tracking table football (foosball) results for everyone in the office. People can register players, log matches, and see a simple leaderboard so rivalries and streaks stay visible without spreadsheets.

## Problem

Match results and bragging rights live in chat threads or memory. There is no single place to see who plays whom, scores, and who is “up” over time.

## Goals

1. **Single deployable unit** — One Docker image runs anywhere (laptop, NAS, small VM) with no separate database container required for MVP.
2. **Low friction** — Add a player and log a match in a few clicks; no training manual.
3. **Honest stats** — Leaderboard reflects logged matches (wins, losses, goals where applicable), not hand-tuned numbers.

## Success criteria

| Criterion        | Definition                                                                                                  |
| ---------------- | ----------------------------------------------------------------------------------------------------------- |
| **Portable**     | `docker run` (with documented env/volume) starts the full app and persists data across restarts.            |
| **Usable**       | A new user can create players, record a match, and see the leaderboard without docs.                        |
| **Maintainable** | API contract is defined in OpenAPI; frontend uses generated types and TanStack Query against that contract. |

## Target users

- **Office players** — Log games and check standings.
- **Whoever hosts it** — Runs Docker and optionally backs up the SQLite file.

## MVP scope

### Must have

- **Players** — Create and list players (name; stable id).
- **Matches** — Record a match with team scores and optional per-side players (supports 1v1 when extras are empty; 2v2 when all slots filled).
- **Leaderboard** — Read-only view derived from matches (e.g. wins, losses, goal difference)—exact columns can follow implementation but must be understandable at a glance.
- **Web UI** — SvelteKit SPA served by the same process as the API (single origin in production).

### Nice to have (only if trivial)

- Expose OpenAPI document over HTTP for tooling (`GET` spec YAML/JSON).
- Clear empty states (“No players yet”) and basic validation errors in the UI.

### Out of scope for MVP

- Accounts, login, or roles (office LAN / trust model).
- Tournaments, seasons, or complex scheduling.
- Mobile-native apps.
- Replacing SQLite with Postgres (revisit if usage grows).

## Technical constraints (non-negotiable for this product)

- **Backend:** Go, REST API described by **OpenAPI 3**, server wiring aligned with generated code (**Echo** + **oapi-codegen**).
- **Persistence:** SQLite (file path configurable via environment).
- **Frontend:** SvelteKit SPA; API client generated from the same OpenAPI spec (**hey-api** / openapi-ts) with **TanStack Query** for server state.
- **Packaging:** Multi-stage build producing **one** Docker image; Go binary embeds built static assets.

## Risks and mitigations

| Risk                         | Mitigation                                                                                                           |
| ---------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| Spec drift between Go and TS | Single `openapi.yaml` in repo; codegen in build/Makefile.                                                            |
| Data loss                    | Document volume mount for DB path; optional backup copy of SQLite file.                                              |
| “Wrong” stats rules          | MVP uses simple rules (win/loss from team scores + player slots); document assumptions in UI or README if ambiguous. |

## Open decisions (defaults)

- **Stats rules:** A player gets a win/loss if they appear on the winning/losing side of a logged match; goals contribute to team totals already stored on the match.
- **Auth:** None for v1.

## Implementation update (2026-03-24)

- Completed the highest-priority remaining MVP task in the web UI: users can now create players and record matches directly from the dashboard.
- Added inline validation and feedback for both forms (required fields, score/date checks, duplicate-player prevention).
- Connected successful writes to TanStack Query cache invalidation so players, matches, and leaderboard views refresh immediately.

---

_This PRD aligns with the engineering plan: OpenAPI-first stack, Echo, single Docker image, SQLite, SvelteKit + hey-api + TanStack Query._
