
# General Idea

This is one unified web app. The primary experience is the **Daily** game — a persistent, async format where users in groups answer one global question per day and guess who wrote each answer. Logging in lands you directly in the daily view.

A secondary mode, **Sessions**, allows ad-hoc real-time games in a JackBox-style environment. Sessions are accessible from a dropdown menu in the app header. They are fully ephemeral — nothing persists when a session ends.

# App Modes

## Daily (Primary)

Login required. Persistent state. Users belong to groups, answer one global question per day, and guess their group members' answers async once the whole group has answered. Scores accumulate over time. Groups have per-day chat threads.

See [Daily Game Flow](#daily-game-flow) for full rules.

## Sessions (Secondary)

No login required. Fully in-memory. A JackBox-style real-time game spun up on demand. Players join with a short code and pick a name — or, if already logged in, their profile name and avatar are pre-populated automatically. Session scores and results are never written to any persistent store.

See the existing phase specs in `docs/` for session game rules.

# Infrastructure

## Authentication

Authentication applies to the **Daily** mode only. Sessions are auth-optional.

Users authenticate via **OAuth only** (Google and/or GitHub). No passwords. Supabase handles the OAuth flow and issues JWTs.

The Go backend verifies Supabase JWTs on authenticated requests. User identity is extracted from the JWT claims (user ID, email, display name, avatar URL).

For sessions, the JWT is read opportunistically on join: if present, the player's display name and avatar are pre-populated from their profile. If absent, the player picks a temp name. Either way, no session data is written to Postgres.

Supabase is the auth and database provider **for now**, with the expectation that the app can migrate to self-hosted Postgres later. The Go backend owns all business logic and writes to Postgres directly (via `pgx/v5`), so Supabase is infrastructure rather than a hard dependency.

## Database

**Supabase Postgres** for the daily game only. The Go backend connects via `pgx/v5` directly — not via Supabase's auto-generated REST APIs. Sessions use the existing in-memory store with no database involvement.

### Schema

```
users
  id            uuid  PK  (matches Supabase auth.users.id)
  email         text  UNIQUE NOT NULL
  display_name  text  NOT NULL
  avatar_url    text
  is_admin      bool  DEFAULT false
  created_at    timestamptz DEFAULT now()

groups
  id            uuid  PK  DEFAULT gen_random_uuid()
  name          text  NOT NULL
  invite_code   text  UNIQUE NOT NULL  -- short code, e.g. XYZ123
  created_by    uuid  FK users.id
  created_at    timestamptz DEFAULT now()

group_members
  group_id      uuid  FK groups.id
  user_id       uuid  FK users.id
  joined_at     timestamptz DEFAULT now()
  PRIMARY KEY (group_id, user_id)

daily_questions
  id            uuid  PK  DEFAULT gen_random_uuid()
  question      text  NOT NULL
  active_date   date  UNIQUE NOT NULL  -- the calendar day this question is served
  created_by    uuid  FK users.id      -- admin who queued it
  created_at    timestamptz DEFAULT now()

daily_answers
  id            uuid  PK  DEFAULT gen_random_uuid()
  question_id   uuid  FK daily_questions.id
  user_id       uuid  FK users.id
  answer_text   text  NOT NULL
  submitted_at  timestamptz DEFAULT now()
  UNIQUE (question_id, user_id)         -- one answer per user per day

daily_guesses
  id              uuid  PK  DEFAULT gen_random_uuid()
  question_id     uuid  FK daily_questions.id
  group_id        uuid  FK groups.id
  guesser_id      uuid  FK users.id
  answer_id       uuid  FK daily_answers.id
  guessed_user_id uuid  FK users.id   -- who the guesser thinks wrote this answer
  submitted_at    timestamptz DEFAULT now()
  UNIQUE (question_id, group_id, guesser_id, answer_id)

group_chat_messages
  id            uuid  PK  DEFAULT gen_random_uuid()
  group_id      uuid  FK groups.id
  question_id   uuid  FK daily_questions.id  -- scopes message to a daily thread
  user_id       uuid  FK users.id
  message       text  NOT NULL
  created_at    timestamptz DEFAULT now()
```

Scores are computed on read (correct guesses per user per day per group) rather than stored.

## Real-time

SSE is used for both modes. For the daily game, per-group SSE streams push:

- Answer submission progress (count only — anonymous until guessing unlocks)
- Guessing unlocked (fires when all group members have answered)
- New chat messages in a group's daily thread

The existing per-lobby SSE infrastructure for sessions is unchanged.

# Daily Game Flow

## Groups

A user creates a group by giving it a name. The system assigns a short invite code. Others join by entering the code. Groups are persistent and accumulate history across days. A user can be in multiple groups simultaneously.

## Daily Question

One global question is served per calendar day, the same for all groups. Questions are admin-curated from a queue (`daily_questions` table, keyed by `active_date`). An admin panel — gated by `is_admin` on the users table — allows queueing future questions.

A user sees today's question immediately on the daily home screen.

## Answering

Each user submits **one answer** per day. That answer is recorded once and applies to all their groups. Answers cannot be changed after submission.

After submitting, the user sees per-group progress: how many members have answered (count only — answers stay anonymous until guessing).

## Guessing

When **all members of a group** have submitted answers, guessing unlocks for that group. Each member independently assigns all anonymous answers to group members.

- Assignments can be changed freely before submitting guesses.
- Once submitted, guesses are locked for that group.
- A user's own answer is included in the pool, but correct self-identification scores no points.

## Scoring

After submitting guesses:
- The user sees which answers they got right and which they missed.
- Score = number of correct assignments (excluding own answer).
- A per-group leaderboard shows cumulative scores across all days.

## Chat

Each group has a per-day chat thread that unlocks when guessing opens for that group. Messages are persistent. Older daily threads remain accessible as history.

# Code Structure

## Backend

```
backend/
  main.go                  -- wires both route trees, shared middleware
  internal/
    sessions/              -- real-time game; existing code moved here
      handler.go           -- Handler struct + shared helpers
      lobby.go
      game.go
      sse.go
      store.go             -- Store interface (lobby, game, token, SSE ops)
      memory.go            -- MemoryStore implementation
      models.go            -- Lobby, Game, Player
    daily/                 -- new; daily game lives entirely here
      handler.go           -- DailyHandler struct
      groups.go
      questions.go
      answers.go
      guesses.go
      chat.go
      sse.go
      store.go             -- DailyStore interface
      postgres.go          -- PostgresStore implementation
      models.go            -- User, Group, DailyQuestion, Answer, Guess, ChatMessage
    auth/                  -- shared auth utilities
      jwt.go               -- Supabase JWT verification
      middleware.go        -- Required and Optional middleware variants
    httputil/              -- shared HTTP helpers
      http.go              -- writeJSON, sseEvent
```

Route structure in `main.go`:
```
/api/sessions/*   → sessions handler (existing routes, nearly unchanged)
/api/daily/*      → daily handler (new)
/api/admin/*      → admin handler (daily, gated by is_admin)
```

## Frontend

```
ui/frontend/src/
  pages/
    daily/                 -- primary experience (logged-in home)
      Dashboard.tsx        -- groups overview, today's question
      Group.tsx            -- group detail: progress, guessing, results
      Chat.tsx             -- daily thread for a group
      Admin.tsx            -- question queue management
    sessions/              -- existing pages moved here
      Home.tsx
      Lobby.tsx
      Game.tsx
  components/
    daily/                 -- daily-specific components
    sessions/              -- session-specific components
    ui/                    -- shadcn components (shared)
  api/
    daily.ts               -- daily API client + types
    sessions.ts            -- existing client.ts renamed
    useSSE.ts              -- shared SSE hook
  types/
    daily.ts
    sessions.ts
```

The app shell (header, nav, auth state) wraps all routes. The header includes a "Create Session" option in a dropdown that routes to `pages/sessions/Home.tsx`.

# Migration Path off Supabase

1. Replace Supabase Auth with a self-hosted OAuth provider (e.g. Keycloak, or `golang.org/x/oauth2` storing tokens in Postgres).
2. Point the `pgx/v5` connection string at a self-hosted Postgres instance.

No application logic changes required — the Go backend never calls Supabase APIs directly.
