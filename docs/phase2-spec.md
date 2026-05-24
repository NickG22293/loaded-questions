# Phase 2 — Question Answering: Detailed Spec

## Overview

After the Asker submits their question (end of Phase 1), the game transitions to Phase 2. All players except the Asker must submit a written answer to the question. Phase 2 ends when every eligible Answerer has submitted.

---

## Rules

### Who answers

- The **Asker does not answer** their own question. They sit out Phase 2.
- All other players are **Answerers**. The denominator for the submission counter is `total players − 1`.

### Answer submission

- Answerers can **edit and resubmit** their answer at any time before Phase 2 ends. The second submission replaces the first.
- A player is counted as "submitted" from the moment of their first submission, even if they later edit it.
- Empty/blank submissions are **not allowed** — the submit button is disabled until the text field has non-whitespace content.

### Anonymity

- Answers are fully anonymous to the **Asker** throughout Phase 2 and Phase 3 (until the Asker assigns them).
- Answerers can see their own answer after submitting but cannot see anyone else's.

### Phase transition

- Phase 2 ends automatically when every Answerer has submitted at least one answer.
- The backend transitions the round to **Phase 3 (ASSIGNING)** and broadcasts `phase_changed`.

---

## Timer

- The lobby creator sets a countdown timer (seconds) before the game starts.
- The timer is **visible** to all players during Phase 2 but **does not auto-advance** the phase when it expires.
- When the timer hits 0 it turns **red and pulses** as a visual warning. Phase 2 continues until all answers are in.

---

## Player views

### Answerer view

| Element | Detail |
|---------|--------|
| Question | Displayed prominently at the top |
| Answer input | Multi-line textarea; placeholder copy encourages honest answers |
| Submit button | Disabled until input is non-empty. Label reads **"Submit Answer"** before first submit, **"Update Answer"** on subsequent edits |
| Submission status | After first submit: confirmation message "Your answer is in!" with option to edit |
| Counter | `X / N answered` — shown below the input, updates live via SSE |
| Timer | Countdown shown; turns red and pulses at 0 |

### Asker view

| Element | Detail |
|---------|--------|
| Question | The question they asked, displayed prominently |
| Counter | `X / N answered` — updates live via SSE |
| Player list | All Answerers listed with a **checkmark** next to names that have submitted; no answer text shown |
| Timer | Same countdown; turns red and pulses at 0 |
| Note | Clear label explaining they are sitting this one out |

---

## SSE events

### `answer_count`

Broadcast to all clients whenever an Answerer submits or resubmits.

```json
{ "submitted": 2, "total": 4, "submittedPlayerIds": ["p1", "p3"] }
```

`total` is `number of players − 1` (excludes the Asker).

`submittedPlayerIds` is included so the Asker's player list can show per-player checkmarks. Anonymity only applies to answer *text* — the act of submitting is not secret.

### `phase_changed`

Broadcast when the last Answerer submits. The payload is the updated `Round` object with `phase: "ASSIGNING"` and the full (server-side) answers array.

---

## Backend changes

### `SubmitAnswer` handler

- Look up game via lobby ID (via `getGameForLobby`).
- Validate phase is `ANSWERING` and player is not the Asker.
- **Append or replace**: if the player has an existing answer in `CurrentRound.Answers`, replace it; otherwise append a new `Answer`.
- Broadcast `answer_count` with the new submitted count and total.
- If submitted count equals total eligible players, transition to `ASSIGNING` and broadcast `phase_changed`.

### `Answer` model

No changes needed. `Answer.PlayerID` is always stored so the backend can do exact-match guessing in Phase 3. It is never sent to the Asker during Phase 2/3.

---

## Frontend changes

### `Game.tsx`

- Add Phase 2 rendering branch (`phase === 'ANSWERING'`).
- `AnswererView`: question, textarea, submit/update button, live counter, timer.
- `AskerView`: question, live counter, player list with submitted indicators, timer.

### Timer component

- Accepts `totalSecs` and `startedAt` (ISO timestamp set when phase begins).
- Counts down to 0, then holds at 0 with red + pulse styling.

### `useSSE` / local SSE in `Game.tsx`

- Handle `answer_count` event: update local `{ submitted, total }` state.
- Handle `phase_changed` event: already wired — updates `currentRound`, which switches the phase branch.

### `client.ts`

- `submitAnswer` already exists and points to the correct endpoint. No changes needed.
