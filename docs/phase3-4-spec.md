# Phase 3 & 4 — Assignment and Scoring: Detailed Spec

## Decisions from Q&A

| Topic | Decision |
|-------|----------|
| Assignment interaction | Click answer to select (highlight), then click player name to assign |
| Live updates | Immediate SSE push per assignment (`assignment_updated` event) |
| Who scores | Only the Asker — 1 point per correct assignment |
| Phase 4 flow | Round results + scoreboard; **any** player can click "Next Round" (first click wins) |
| Game over | Dedicated GAME_OVER phase/screen with winner announcement and final scoreboard |

---

## Phase 3 — ASSIGNING

### Who acts
- Only the **Asker** can assign answers.
- **Answerers** watch live as each assignment happens.

### Assignment flow (click-to-select → click player)
1. Asker sees all answer texts in a list — anonymous (no author names shown).
2. Asker clicks an answer card → it becomes **selected** (highlighted blue border). Clicking a selected answer deselects it.
3. With an answer selected, Asker clicks a player name → `POST /lobbies/:id/assign` is called.
4. The answer is now assigned; the player name appears next to that answer card.
5. Any assignment can be **changed** before Lock In: select the answer again and click a different player.
6. The Asker cannot assign an answer to themselves (they didn't answer).

### Validation
- Phase must be `ASSIGNING`.
- Caller must be the Asker.
- `answerId` must reference an existing answer in `CurrentRound.Answers`.
- `playerId` must exist in `game.Players` and must not be the Asker.

### Live SSE: `assignment_updated`
Broadcast to all clients after each `AssignAnswer` call.  
Payload: the updated `Round` object (answers now have `assignedTo` populated).

### Lock In
- "Lock In" button on the Asker's view is **disabled** until every answer has a non-empty `AssignedTo`.
- `POST /lobbies/:id/lock` → validates all assigned, calculates score, transitions to `SCORING`, broadcasts `phase_changed` with full `Game`.

### Score calculation (inside LockAssignments)
```
score += 1  for each answer where answer.AssignedTo == answer.PlayerID
game.Players[askerIdx].Score += score
game.CurrentRound.Phase = SCORING
```

---

## Phase 3 — Answerer waiting view

- "The Asker is reviewing your answers…" with a spinner initially.
- As assignments arrive via SSE, a list grows showing: `"[Answer text]" → [Player Name]`
- The row assigned TO the current player is highlighted.

---

## Phase 4 — SCORING

### All players see

- The question at the top.
- Each answer as a card:
  - Answer text
  - "Asker guessed → [Name]" (who Asker assigned it to)
  - "Actually → [Name]" (who really wrote it — now revealed)
  - ✓ correct / ✗ wrong indicator
- Asker's round score delta: **"+N point(s)!"**
- Full scoreboard: all players with total scores.

### "Next Round" button
- Visible to **all** players.
- First player to click calls `POST /lobbies/:id/next`.
- Backend: checks for winner → if no winner, creates next round.

### Next Asker rotation
Round-robin by `game.Players` order:
```
nextAskerIdx = (currentAskerIdx + 1) % len(game.Players)
```

---

## GAME_OVER

### Trigger
After Phase 4, if any player's score `>= game.TargetScore`, the game ends.

### What happens
- `game.WinnerID` is set.
- `game.CurrentRound.Phase` is set to `GAME_OVER`.
- Broadcasts `phase_changed` with full `Game`.

### Screen
- All players: "🎉 [Winner Name] wins!"
- Final scoreboard.
- "Back to Home" button.

---

## SSE Events (updated contract)

### `phase_changed` — payload changed
Now carries the **full `Game` object** (not just `Round`) so clients receive updated player scores and game state atomically.

```json
{ "id": "...", "lobbyId": "...", "players": [...], "currentRound": {...}, ... }
```

### `assignment_updated` — new
Carries the updated `Round` (mid-phase, no score changes yet).

```json
{ "roundNumber": 1, "askerId": "...", "answers": [{"id":"a1","text":"...","assignedTo":"p2",...}], ... }
```

---

## Backend changes summary

| Handler | Change |
|---------|--------|
| `SubmitQuestion` | Broadcast `phase_changed` with full `game` (was just `currentRound`) |
| `SubmitAnswer` | Broadcast `phase_changed` with full `game` |
| `AssignAnswer` | Full implementation — set `AssignedTo`, broadcast `assignment_updated` |
| `LockAssignments` | Score calculation, phase → SCORING, broadcast `phase_changed` with game |
| `NextRound` | Any player; check winner → GAME_OVER or rotate Asker → ASKING |

## Frontend changes summary

| File | Change |
|------|--------|
| `types/index.ts` | Add `'GAME_OVER'` to Phase |
| `Game.tsx` | `phase_changed` parses full `Game`; add `assignment_updated` handler; add ASSIGNING, SCORING, GAME_OVER branches |
