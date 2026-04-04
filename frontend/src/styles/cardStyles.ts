import type React from 'react';
import { EXPANSION_GAP_PX } from './motionPresets';

/** Tailwind classes for focus-visible ring on card selection buttons. */
export const focusRingCard =
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ds-accent focus-visible:ring-offset-1 focus-visible:ring-offset-transparent rounded-lg';

/** Tailwind classes for hover feedback on clickable cards (non-AnimatedCard). */
export const hoverCardClass = 'cursor-pointer transition-[transform,box-shadow] duration-150';

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
