# ADR 0024: Fluid & Tactile UI Redesign

## Status

Accepted

## Date

2026-03-21

## Context

The frontend uses pure Tailwind CSS v4 with only 3 CSS keyframe animations and no animation, gesture, or sound libraries. While functional, the UI feels plain and lacks the polish expected of a modern card game application. Users on mobile devices have limited tactile feedback and the visual design does not convey the premium feel of a card game.

Epic #788 proposes a phased UI redesign across 4 sub-issues (#784–#787) to transform the UI into a "Fluid & Tactile" experience.

## Decision

We will implement the redesign in 4 incremental PRs:

1. **Design System & Theme (#784)**: Glassmorphism design tokens and utility classes using pure CSS (no new dependencies). Glass panels replace opaque backgrounds for a modern, layered look. System font stacks via CSS variables for typography. `@supports` fallback for browsers without `backdrop-filter`.

2. **Physics-based Fluid Animation (#785)**: Add `framer-motion` (~33KB gzipped) for spring-based card deal/select animations, phase transitions, and win celebrations. All animations respect `prefers-reduced-motion`. Globally mocked in tests.

3. **Ergonomic Mobile Layout (#786)**: Restructure game pages to place player hand cards in a floating footer within the thumb zone. No new dependencies.

4. **Gesture & Haptics/Sound (#787)**: Add `@use-gesture/react` (~3KB) for swipe/tap gestures and `use-sound`+`howler` (~11KB) for audio feedback. Small OGG sound assets (~25KB total). `navigator.vibrate()` for haptic feedback with graceful degradation.

### Technology choices

- **framer-motion** over CSS animations: Spring physics, `AnimatePresence` for exit animations, `layoutId` for shared layout transitions — capabilities not achievable with CSS alone.
- **@use-gesture/react** over raw pointer events: Declarative gesture API that pairs naturally with framer-motion's spring system.
- **use-sound/howler** over Web Audio API: Simple API for fire-and-forget sound effects with format fallback and mobile compatibility.
- **No external fonts**: System font stacks avoid additional network requests and FOIT/FOUT.

### Bundle budget

Total addition: ~72KB gzipped (33 + 3 + 11 + 25 for assets). All libraries are tree-shakeable.

## Consequences

### Positive

- Modern, polished card game experience with glassmorphism, smooth animations, and tactile feedback
- Improved mobile usability with thumb-zone optimized layout and gesture support
- Accessible by default: `prefers-reduced-motion` respected, button/keyboard alternatives for all gestures
- Incremental delivery: each PR is independently shippable and testable

### Negative

- ~72KB bundle size increase (acceptable for a game application)
- framer-motion becomes a significant dependency; future removal would require substantial refactoring
- Sound assets require CC0-licensed source files to be sourced and committed
- Glass panels require `@supports` fallback testing for older browsers
