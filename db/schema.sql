-- Run this once in Supabase's SQL editor (or via psql) to create the daily game schema.
-- See docs/follow-ups.md: proper migration tooling (golang-migrate) is a follow-up.

CREATE TABLE IF NOT EXISTS users (
    id           UUID PRIMARY KEY,
    email        TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    avatar_url   TEXT,
    is_admin     BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS groups (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL,
    invite_code  TEXT UNIQUE NOT NULL,
    created_by   UUID NOT NULL REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS group_members (
    group_id   UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS daily_questions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question    TEXT NOT NULL,
    active_date DATE UNIQUE NOT NULL,
    created_by  UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One answer per user per question (global, not per-group).
CREATE TABLE IF NOT EXISTS daily_answers (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id  UUID NOT NULL REFERENCES daily_questions(id),
    user_id      UUID NOT NULL REFERENCES users(id),
    answer_text  TEXT NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (question_id, user_id)
);

-- One set of guesses per user per group per question. Locked after submit.
CREATE TABLE IF NOT EXISTS daily_guesses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id     UUID NOT NULL REFERENCES daily_questions(id),
    group_id        UUID NOT NULL REFERENCES groups(id),
    guesser_id      UUID NOT NULL REFERENCES users(id),
    answer_id       UUID NOT NULL REFERENCES daily_answers(id),
    guessed_user_id UUID NOT NULL REFERENCES users(id),
    submitted_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (question_id, group_id, guesser_id, answer_id)
);

-- Per-group chat scoped to a daily question thread.
CREATE TABLE IF NOT EXISTS group_chat_messages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES daily_questions(id),
    user_id     UUID NOT NULL REFERENCES users(id),
    message     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_group_members_user    ON group_members(user_id);
CREATE INDEX IF NOT EXISTS idx_daily_answers_question ON daily_answers(question_id);
CREATE INDEX IF NOT EXISTS idx_daily_guesses_guesser ON daily_guesses(guesser_id, group_id, question_id);
CREATE INDEX IF NOT EXISTS idx_chat_group_question   ON group_chat_messages(group_id, question_id, created_at);
