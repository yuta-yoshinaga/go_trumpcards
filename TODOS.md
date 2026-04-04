# TODOS

## Game Feel

### Performance benchmark for AnimatedPile cascade
**Priority:** P2
**What:** After solitaire integration, benchmark 20+ card auto-complete on mobile for GPU jank.
**Why:** Spider Solitaire has 104 cards. Even with batching (4-6 per group), the cascade involves many sequential animation groups. Mobile GPU compositing could cause jank.
**Fallback:** CSS `transform` with `will-change: transform` if spring animations jank on mobile.
**Depends on:** Game feel toolkit shipped and solitaire games using AnimatedPile in production.

### useGameAtmosphere hook (v2 game feel)
**Priority:** P3
**What:** Build a hook that ties animation/sound parameters to game state transitions. E.g., springs tighten when close to busting in Blackjack, background subtly darkens during tense moments.
**Why:** Transforms game feel from good to exceptional. Unique in the React card game space.
**Depends on:** Game feel toolkit (Phases 1-4) shipped and validated with real users.
**Reference:** Design doc at `~/.gstack/projects/yuta-yoshinaga-go_trumpcards/yuta-develop-design-20260404-151402.md`, "Future Directions" section.

## Completed

### Replace placeholder sound files with real recordings
**Priority:** P1
**Completed:** 2026-04-04
**What:** Sourced 10 CC0 sound files from Kenney.nl casino audio pack and OpenGameArt.org. Card sounds (deal, flip, select, place, shuffle) from Kenney, UI sounds (chip-click, error-buzz, turn-tick) from Kenney, win fanfare from OpenGameArt (winfretless), loss trumpet from OpenGameArt (losetrumpet). Total ~144KB OGG.
