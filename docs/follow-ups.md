# Follow-ups

Items deferred from completed work, in roughly priority order.

---

## Auth layer

- **Wire `Optional` middleware into session join** — If a logged-in user joins a session, pre-populate their display name and avatar from the JWT profile. Deferred from the auth PR; do this alongside the first daily route wiring.

---

## Daily backend

- **Migration tooling** — Add `golang-migrate` (or Goose) with numbered files under `db/migrations/`. For now the schema is in `db/schema.sql` and run manually via Supabase's SQL editor.

- **Admin routes** — `GET/POST/PUT/DELETE /api/admin/questions` for queueing daily questions. Gated by `is_admin` on the `users` table. Implement once the core daily routes are live.

- **Integration tests for `PostgresStore`** — Handler tests mock `DailyStore`. The Postgres implementation itself has no test coverage until a Docker Compose dev environment with a real Postgres instance is added.

- **Timezone handling** — The 6pm guessing deadline is currently server-local time. If the user base is multi-timezone, this needs revisiting (e.g. fixed UTC offset, or per-group timezone setting).

- **Answer edit window** — Spec says answers cannot be changed after submission. If that ever relaxes (e.g. a short grace period), the `SubmitAnswer` store method needs an upsert-with-window strategy.

---

## Sessions

- **Session persistence for logged-in users** — Currently session game data is never written to Postgres, even for logged-in users. If per-user session history is ever wanted, a session summary store would be needed.

---

## Infrastructure

- **Docker Compose dev environment** — A `docker-compose.yml` with a local Postgres for running integration tests and daily routes locally without a live Supabase project.

- **Helm chart daily env vars** — `DATABASE_URL`, `SUPABASE_JWKS_URL`, and `SUPABASE_JWT_ISSUER` need adding to the Helm chart values/secrets.
