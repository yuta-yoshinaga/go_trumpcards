import type React from 'react';
import { EXPANSION_GAP_PX } from './motionPresets';

/**
 * Classes for the keyboard focus indicator on card selection buttons.
 *
 * The indicator itself is a stylesheet rule (`.card-focus-ring:focus-visible`
 * in `index.css`), deliberately not a Tailwind utility. Card buttons set inline
 * styles on **both** channels a utility could use:
 *
 * - `boxShadow` — `selectedCardStyle` / `highlightCardStyle` /
 *   `playableCardStyle` / `smartHighlightStyle`; the "off" branch sets `'none'`,
 *   which still counts as an inline declaration.
 * - `outline` — `trumpRingStyle` / `meldCardStyle` / `playableRingStyle`;
 *   `meldCardStyle` sets one on *every* hand card during the GinRummy and
 *   Chinchon discard and layoff phases.
 *
 * Inline styles beat class-based declarations, so both a `ring-*` and a plain
 * `outline-*` utility are erased — a `ring` everywhere, an `outline` in exactly
 * the phases a player is most likely to be navigating by keyboard. The
 * stylesheet rule carries `!important` so the indicator survives regardless.
 * See issue #5359.
 *
 * Any new decorative helper that sets `outline` or `boxShadow` inline is fine;
 * the focus indicator no longer competes with them.
 */
export const focusRingCard = 'rounded-lg card-focus-ring';

/** Tailwind classes for hover feedback on clickable cards (non-AnimatedCard). */
export const hoverCardClass = 'cursor-pointer transition-[transform,box-shadow] duration-150';

/**
 * Transparent placeholder border for cards that are never selectable but
 * sit next to selectable ones. Keeps their box-model width identical to a
 * `selectedCardStyle(false)` card so the layout doesn't jump when selection
 * toggles in adjacent rows. Single source of truth for the unselectable
 * "no visual feedback" state.
 */
export const placeholderCardStyle: React.CSSProperties = { border: '3px solid transparent' };

/** Return inline styles for a card with selection highlight and lift effect. */
export function selectedCardStyle(isSelected: boolean): React.CSSProperties {
  return {
    border: isSelected ? '3px solid var(--color-game-card-selected)' : '3px solid transparent',
    transform: isSelected ? 'translateY(-8px)' : 'none',
    transition: 'transform 0.15s, border 0.15s, box-shadow 0.15s',
    boxShadow: isSelected ? '0 4px 12px rgba(59, 130, 246, 0.4), 0 0 20px rgba(59, 130, 246, 0.15)' : 'none',
  };
}

/** Return inline styles for a card with playable border highlight. */
export function playableCardStyle(isPlayable: boolean): React.CSSProperties {
  return {
    border: isPlayable ? '3px solid var(--color-game-status-active)' : '3px solid transparent',
    boxShadow: isPlayable ? '0 0 8px rgba(92, 184, 92, 0.3)' : 'none',
  };
}

/**
 * Return inline styles for a card highlighted as an actionable option (e.g. an
 * exposable card in Gong Zhu). Uses the warning accent so it reads distinctly
 * from the blue selection border. Applied via inline style (not a Tailwind
 * `ring`) because the selection styles set `boxShadow` inline, which would
 * otherwise override a class-based ring.
 */
export function highlightCardStyle(): React.CSSProperties {
  return {
    border: '3px solid var(--color-ds-warning)',
    boxShadow: '0 0 8px rgba(232, 146, 58, 0.45)',
    transition: 'border 0.15s, box-shadow 0.15s',
  };
}

/**
 * Return inline styles for a subtle "trump" marker ring. Uses `outline`
 * (not `border`/`boxShadow`) so it stacks additively on top of the selection
 * and playable borders without clobbering them — a trump card can be selected
 * or restricted and still show its ring. Applied e.g. to Doppelkopf trumps.
 */
export function trumpRingStyle(): React.CSSProperties {
  return {
    outline: '2px solid var(--color-ds-accent)',
    outlineOffset: '1px',
    borderRadius: 8,
  };
}

/**
 * Return inline styles marking a card the hint is pointing at.
 *
 * Uses `outline` (not `border`/`boxShadow`) for the same reason as
 * {@link trumpRingStyle}: hand cards carry an inline `boxShadow` from
 * {@link selectedCardStyle}, which would override a Tailwind `ring-*` class —
 * including the `none` it sets while unselected. An outline stacks on top of
 * the selection border instead of fighting it, so a suggested card can also be
 * selected.
 */
export function hintRingStyle(): React.CSSProperties {
  return {
    outline: '2px solid var(--color-ds-warning)',
    outlineOffset: '1px',
    borderRadius: 8,
  };
}

/**
 * Return inline styles that color-code a hand card by meld membership (e.g. Gin
 * Rummy): a melded card gets a green outline, a deadwood card a subtle grey one.
 * Uses `outline` (not `border`/`boxShadow`) so it stacks additively on top of
 * the selection border and lift without clobbering them — a card can be both
 * melded and selected. Purely decorative; never blocks clicks.
 */
export function meldCardStyle(isMelded: boolean): React.CSSProperties {
  return {
    outline: isMelded ? '2px solid var(--color-ds-success)' : '2px dashed rgba(148, 163, 184, 0.5)',
    outlineOffset: '1px',
    borderRadius: 8,
  };
}

/**
 * Return inline styles for an additive "playable right now" success ring.
 * Uses `outline` (not `border`/`boxShadow`) so it stacks on top of the
 * selection border and any inline `boxShadow` without clobbering them — a
 * playable card can also be selected and still show its ring. Uses the
 * `--color-ds-success` design token. Applied e.g. to Speed hand cards.
 */
export function playableRingStyle(): React.CSSProperties {
  return {
    outline: '2px solid var(--color-ds-success)',
    outlineOffset: '1px',
    borderRadius: 8,
  };
}

/**
 * Return adjusted overlap margin for a card that is adjacent to a selected card on mobile.
 * Reduces the negative overlap by EXPANSION_GAP_PX, effectively widening the visible area.
 * Returns the original overlap when the card is not adjacent to a selection.
 */
export function expansionMargin(isNeighborOfSelected: boolean, baseOverlap: number): number {
  if (!isNeighborOfSelected) return baseOverlap;
  // Reduce the negative overlap (make it less negative = more visible area)
  return baseOverlap + EXPANSION_GAP_PX;
}

/** Return inline styles combining playable border + enhanced glow for thumb-zone visibility. */
export function smartHighlightStyle(isPlayable: boolean): React.CSSProperties {
  return {
    border: isPlayable ? '3px solid var(--color-game-status-active)' : '3px solid transparent',
    boxShadow: isPlayable ? '0 0 10px rgba(92, 184, 92, 0.4), 0 0 20px rgba(92, 184, 92, 0.15)' : 'none',
    transition: 'border 0.15s, box-shadow 0.15s',
  };
}
