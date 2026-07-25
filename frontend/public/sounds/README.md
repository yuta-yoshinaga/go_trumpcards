# Sound Assets

Source of truth for the web GUI's sound effects, played via
`src/providers/SoundProvider.tsx` (Howler). The repo-root `public/sounds/`
copies are Vite build output (`vite.config.ts` `outDir: '../public'`) — never
edit those by hand.

## Files

| Sound | Files | Usage |
|-------|-------|-------|
| cardDeal | `card-deal.{ogg,mp3}` | Per-card deal animation (AnimatedCard) |
| cardFlip | `card-flip.{ogg,mp3}` | Card flip animation (AnimatedCardBack) |
| cardSelect | `card-select.{ogg,mp3}` | Card selection |
| cardPlace | `card-place.{ogg,mp3}` | Generic action feedback (central useGameApi tap) |
| shuffle | `shuffle.{ogg,mp3}` | Deal / redeal (`reset` command) |
| chipClick | `chip-click.{ogg,mp3}` | Bet / call / raise / all-in (BettingControls) |
| winFanfare | `win-fanfare.{ogg,mp3}` | Win celebration (GamePageShell) |
| lossThud | `loss-thud.{ogg,mp3}` | Loss (GamePageShell, pages passing `winShow`) |
| errorBuzz | `error-buzz.{ogg,mp3}` | Displayed errors (ErrorAlert) |
| turnTick | `turn-tick.{ogg,mp3}` | Human-turn start (GamePageShell) |

## mp3 fallbacks

Safari/iOS cannot decode Ogg Vorbis. The `.mp3` files are generated from the
`.ogg` sources — **never edit them independently**. After adding or
re-recording any `.ogg`, regenerate with:

```sh
./scripts/generate-sound-fallbacks.sh   # requires ffmpeg
```

## Licensing

All sound files in this directory must be **CC0 (public domain)** licensed.
When adding or replacing files, document the source and license here.

- Current `.ogg` set: procedurally generated for this project (CC0).
- `.mp3` files: derived from the `.ogg` set via the script above.
