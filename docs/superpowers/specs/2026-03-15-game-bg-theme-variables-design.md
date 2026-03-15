# Game Background Color Theme Variables

**Issue:** [#657](https://github.com/yuta-yoshinaga/go_trumpcards/issues/657)
**Date:** 2026-03-15

## Problem

Game page background colors are hardcoded as Tailwind arbitrary values (`bg-[#xxxxxx]`) across 11 page files and 1 component file (22 source instances + 3 test file references = 25 total). This prevents centralized color management and theme extensibility.

## Solution

Add CSS theme variables to `frontend/src/index.css` and replace all hardcoded `bg-[#...]` classes with Tailwind theme classes. Zero visual change.

## Color Variables

Five distinct color pairs (main + dark footer variant):

| CSS Variable | Hex | Tailwind Class | Games |
|---|---|---|---|
| `--color-game-bg-green` | `#1a5c1a` | `bg-game-bg-green` | OldMaid, Daifugo, Sevens |
| `--color-game-bg-green-dark` | `#163e16` | `bg-game-bg-green-dark` | OldMaid, Daifugo, Sevens (footer) |
| `--color-game-bg-green-bright` | `#008000` | `bg-game-bg-green-bright` | BlackJack |
| `--color-game-bg-green-bright-dark` | `#005a00` | `bg-game-bg-green-bright-dark` | BlackJack (footer) |
| `--color-game-bg-green-poker` | `#1a6b1a` | `bg-game-bg-green-poker` | Poker, Hold'em |
| `--color-game-bg-green-poker-dark` | `#155715` | `bg-game-bg-green-poker-dark` | Poker, Hold'em (footer) |
| `--color-game-bg-casino` | `#0d5016` | `bg-game-bg-casino` | Klondike, Baccarat |
| `--color-game-bg-casino-dark` | `#0a3a10` | `bg-game-bg-casino-dark` | Klondike (footer). Note: Baccarat footer uses `bg-gray-800` intentionally — not included. |
| `--color-game-bg-blue` | `#1a2c5c` | `bg-game-bg-blue` | Doubt, Hearts, Memory |
| `--color-game-bg-blue-dark` | `#101c3a` | `bg-game-bg-blue-dark` | Doubt, Hearts, Memory (footer) |

## Changes

### 1. `frontend/src/index.css` — Add to `@theme` block

```css
--color-game-bg-green: #1a5c1a;
--color-game-bg-green-dark: #163e16;
--color-game-bg-green-bright: #008000;
--color-game-bg-green-bright-dark: #005a00;
--color-game-bg-green-poker: #1a6b1a;
--color-game-bg-green-poker-dark: #155715;
--color-game-bg-casino: #0d5016;
--color-game-bg-casino-dark: #0a3a10;
--color-game-bg-blue: #1a2c5c;
--color-game-bg-blue-dark: #101c3a;
```

### 2. Page files — Replace hardcoded colors

| File | Before | After |
|---|---|---|
| BlackJackPage.tsx | `bg-[#008000]` | `bg-game-bg-green-bright` |
| BlackJackPage.tsx | `bg-[#005a00]` | `bg-game-bg-green-bright-dark` |
| PokerPage.tsx | `bg-[#1a6b1a]` | `bg-game-bg-green-poker` |
| PokerPage.tsx | `bg-[#155715]` | `bg-game-bg-green-poker-dark` |
| HoldemPage.tsx | `bg-[#1a6b1a]` | `bg-game-bg-green-poker` |
| HoldemPage.tsx | `bg-[#155715]` | `bg-game-bg-green-poker-dark` |
| OldMaidPage.tsx | `bg-[#1a5c1a]` | `bg-game-bg-green` |
| OldMaidPage.tsx | `bg-[#163e16]` | `bg-game-bg-green-dark` |
| DaifugoPage.tsx | `bg-[#1a5c1a]` | `bg-game-bg-green` |
| DaifugoPage.tsx | `bg-[#163e16]` | `bg-game-bg-green-dark` |
| SevensPage.tsx | `bg-[#1a5c1a]` | `bg-game-bg-green` |
| SevensPage.tsx | `bg-[#163e16]` | `bg-game-bg-green-dark` |
| KlondikePage.tsx | `bg-[#0d5016]` | `bg-game-bg-casino` |
| KlondikePage.tsx | `bg-[#0a3a10]` | `bg-game-bg-casino-dark` |
| BaccaratPage.tsx | `bg-[#0d5016]` | `bg-game-bg-casino` |
| OldMaidSetupScreen.tsx | `bg-[#1a5c1a]` | `bg-game-bg-green` |
| DoubtPage.tsx | `bg-[#1a2c5c]` | `bg-game-bg-blue` |
| DoubtPage.tsx | `bg-[#101c3a]` | `bg-game-bg-blue-dark` |
| HeartsPage.tsx | `bg-[#1a2c5c]` | `bg-game-bg-blue` |
| HeartsPage.tsx | `bg-[#101c3a]` | `bg-game-bg-blue-dark` |
| MemoryPage.tsx | `bg-[#1a2c5c]` | `bg-game-bg-blue` |
| MemoryPage.tsx | `bg-[#101c3a]` | `bg-game-bg-blue-dark` |

### 3. Test files — Update any hardcoded color assertions

- KlondikePage.test.tsx references `bg-[#0a3a10]` — update to `bg-game-bg-casino-dark`
- GameFooter.test.tsx references `bg-[#005a00]` — update to `bg-game-bg-green-bright-dark`

## Testing

- `npm run build` — Tailwind compiles the new theme classes
- `npm run check` — Biome lint passes
- `npm test` — All unit tests pass (update any color assertions in test files)
- Visual: no change in rendered appearance
