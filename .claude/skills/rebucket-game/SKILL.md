---
name: rebucket-game
description: Move an existing game between Cloudflare Worker size buckets (casino/classic/solo/extra/extra2/extra3) when its current worker exceeds or nears the 1 MB gzip free-tier limit. Scripts the 6 touchpoints (build tags, registry, KV registration, count assertions, gameExec.ts, docs) and measures the result locally with TinyGo. Use when the size-check CI step fails ("EXCEEDS free tier limit") or when routing a game to a different worker (e.g. "/rebucket-game scarto solo", "move <game> to <worker>").
---

# Rebucket Game

Move `<game>` from its current worker bucket to `<target>` (`casino` | `classic` | `solo` |
`extra` | `extra2` | `extra3`). The bucket is a **binary-size bucket, not a user-facing taxonomy** — rebucketing
changes nothing about gameplay or UI, only which TinyGo WASM binary the game compiles into.
Missing any touchpoint ships a broken worker route, so do ALL six, in one commit.

Known precedent: scarto was built for `extra`, pushed it 10 KB over the limit, and was
rebucketed to `solo` (this exact procedure).

## Inputs

- `<game>` — registry key (run `go run ./cmd/trumpcards games --short` to confirm spelling).
- `<target>` — destination bucket. If not given, pick the worker with the most gzip headroom
  (see Size verification below). Current standing guidance: `extra` is FULL (~1.048 MB) and
  `classic` is near the limit — prefer `solo` unless CI numbers say otherwise.

## The 6 touchpoints

1. **Build tags** — retag EVERY production (non-test) `.go` file of the game with
   `//go:build !js || !wasm || <target>` (first line, blank line after). Find them with:
   ```sh
   grep -rln "<Game>" internal/domain internal/usecase internal/adapter --include="*.go" | grep -v _test
   ```
   Typical set: `internal/domain/<Game>*.go`, `internal/usecase/<Game>Interactor.go`,
   `internal/adapter/controller/<Game>{Web,Cui}Controller.go`,
   `internal/adapter/presenter/<Game>{Web,Cui}Presenter.go`. ALL of them need the tag —
   one missed file bloats the old worker AND breaks dead-code elimination.

2. **Registry** — `internal/infrastructure/games/registry.go`: change the game's entry to
   `{Name: "<game>", Category: Category<Target>}`.

3. **KV registration** — move the whole `games.RegisterKVGame("<game>", games.Category<Old>, ...)`
   block from `internal/infrastructure/games/<old>/<old>.go` to
   `internal/infrastructure/games/<target>/<target>.go`, updating the Category argument to
   `games.Category<Target>`. Move any imports it needs.

4. **Count assertions** — `internal/infrastructure/games/registry_test.go`: decrement
   `expected<Old>`, increment `expected<Target>`. `expectedTotal` is derived and must NOT
   change. (The commit hook runs `count-audit.sh` and will block if this is missed.)

5. **Frontend routing** — `frontend/src/api/gameExec.ts` (the `workerUrl` map; it moved out
   of `gameApi.ts` in #4434): change the game's entry to `<game>: WORKER_<TARGET>`.

6. **Docs** — `docs/cloudflare-workers.md`: move the key between the per-worker rows and fix
   both counts. `TestDocsMatchRegistry` enforces this, so a miss fails the suite rather than
   shipping a stale table.

## Use the scripts, not hand edits

```sh
python3 .claude/skills/rebucket-game/scripts/crossrefs.py <target> <game>...  # check FIRST
python3 .claude/skills/rebucket-game/scripts/move-game.py  <target> <game>...
python3 .claude/skills/rebucket-game/scripts/move-game.py  --check            # layout
```

`move-game.py` does touchpoints 1-5 atomically (nothing is written unless every game in the
batch succeeds) and refuses to split an implementation group. Touchpoint 6 is still manual.

## Two traps that are invisible until the build fails

**A game is not always its own implementation.** Several games are variants riding on
another game's domain type -- `razz` and `sevencardstud` share `SevenCardStud`, `spanish21`
rides on `BlackJack`, `irishpoker` on `Pineapple`, and there are three more. They share every
production file, so the bucket belongs to the implementation, not the game. `move-game.py`
refuses a partial group; the full list is `move-game.py --check`'s companion output.

**A bucket's games can be welded together by shared package-level symbols.** `GameResult`
lives in `BlackJack.go` and 19 other casino games use it; `compareHighCardsSlice` lives in
`HoldemPlayer.go` and 12 do. Moving blackjack or holdem out of casino strands up to 82
symbols and the remaining games stop compiling. Treat that count as an upper bound: it is a
lexical match, so a symbol sharing a name with some game's struct field is credited with
users that never call it, and declarations in untagged files are excluded because those
compile into every worker regardless. Run `crossrefs.py` before choosing a move set -- it reports, in both directions,
every package-level symbol that would cross the boundary. As of ADR-0036 Phase 2 only 6 of
casino's 55 units are freely movable, versus 23 of solo's 51.

## Verification (before commit)

```sh
go build ./...
go test -tags test ./internal/infrastructure/games/...   # counts + docs-vs-registry guard
go test -tags test ./cmd/...                             # AllCategories drives the CLI listing
bash .claude/skills/new-game/scripts/count-audit.sh
```

Then run the `game-registration-checker` agent for an independent sweep of all touchpoints.

## Size verification (measure locally first)

TinyGo 0.40.1 + Go 1.25.8 are installed now, and local builds reproduce CI byte for byte
(verified: extra2 = 594139 raw / 232016 gzip in both). Measure before pushing:

```sh
.claude/skills/rebucket-game/scripts/measure.sh extra2          # one worker, ~3.5 min
.claude/skills/rebucket-game/scripts/measure.sh                  # all six, ~21 min
```

Budget ~14.4 KB gzip per average game, but the spread is wide (scarto measured 20,332 B) --
and every worker pays a fixed 232 KB baseline for the Go runtime regardless of its contents,
so never estimate a bucket's capacity as "total gzip / games".

`make` is not installed here; `measure.sh` mirrors the Makefile macro exactly. CI remains the
final word:

```sh
gh run list --workflow=cloudflare-workers-build.yml --branch <branch> --limit 1
gh run watch <run-id>
```

The "Check size limit" step fails (exit 1) if `gzip(app.wasm) > 1048576` bytes for any
worker. Both the OLD worker (should shrink) and the TARGET worker (must stay under 1 MB)
matter. To gauge a bucket's current headroom before choosing `<target>`, read the step
summary of the latest successful run on `develop`:

```sh
gh run list --workflow=cloudflare-workers-build.yml --branch develop --limit 1 --json databaseId \
  --jq '.[0].databaseId' | xargs -I{} gh run view {} --log | grep -A6 "Check size limit" | grep -E "Raw|gzip"
```

If the target worker also overflows, pick the next-roomiest bucket and repeat — do not
split a game's files across buckets.
