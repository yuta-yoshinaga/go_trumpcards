import { describe, expect, it } from 'vitest';
import {
  expansionMargin,
  focusRingCard,
  playableCardStyle,
  selectedCardStyle,
  smartHighlightStyle,
} from './cardStyles';
import { EXPANSION_GAP_PX } from './motionPresets';

describe('selectedCardStyle', () => {
  it('returns selected styles when true', () => {
    const style = selectedCardStyle(true);
    expect(style.border).toBe('3px solid var(--color-game-card-selected)');
    expect(style.transform).toBe('translateY(-8px)');
    expect(style.transition).toBe('transform 0.15s, border 0.15s, box-shadow 0.15s');
    expect(style.boxShadow).toBe('0 4px 12px rgba(59, 130, 246, 0.4), 0 0 20px rgba(59, 130, 246, 0.15)');
  });

  it('returns unselected styles when false', () => {
    const style = selectedCardStyle(false);
    expect(style.border).toBe('3px solid transparent');
    expect(style.transform).toBe('none');
    expect(style.boxShadow).toBe('none');
  });
});

describe('playableCardStyle', () => {
  it('returns playable styles when true', () => {
    const style = playableCardStyle(true);
    expect(style.border).toBe('3px solid var(--color-game-status-active)');
    expect(style.boxShadow).toBe('0 0 8px rgba(92, 184, 92, 0.3)');
  });

  it('returns non-playable styles when false', () => {
    const style = playableCardStyle(false);
    expect(style.border).toBe('3px solid transparent');
    expect(style.boxShadow).toBe('none');
  });
});

describe('focusRingCard', () => {
  // The focus indicator must survive the inline styles every card button sets.
  //
  // Tailwind's `ring-*` compiles to `box-shadow`, and each card button sets
  // `boxShadow` inline via selectedCardStyle / highlightCardStyle -- including
  // the unselected branch, which sets `'none'`. Inline styles beat class-based
  // declarations, so a ring-based focus indicator is silently dropped in every
  // state. `outline` is a separate property and stacks additively, which is the
  // same reasoning trumpRingStyle already documents. See issue #5359.
  it('uses outline, not a ring, so inline boxShadow cannot erase it', () => {
    expect(focusRingCard).toContain('focus-visible:outline-2');
    expect(focusRingCard).toContain('rounded-lg');
    expect(focusRingCard).not.toMatch(/focus-visible:ring/);
  });

  // Removing the browser default leaves nothing when the ring is overridden,
  // so there must be no `outline-none` to fall back from.
  it('does not disable the browser default outline', () => {
    expect(focusRingCard).not.toContain('outline-none');
  });

  it('is visible against the card, not transparent', () => {
    expect(focusRingCard).toMatch(/focus-visible:outline-(\[var\(--color-ds-accent\)\]|ds-accent)/);
  });
});

describe('focusRingCard vs the inline styles it coexists with', () => {
  // Guards the *interaction* rather than either side alone: this is what makes
  // the bug reproducible, and it is why asserting on focusRingCard's string in
  // isolation was not enough to catch it before.
  it('every card style helper that sets boxShadow leaves outline untouched', () => {
    for (const style of [
      selectedCardStyle(true),
      selectedCardStyle(false),
      playableCardStyle(true),
      playableCardStyle(false),
    ]) {
      expect(style).toHaveProperty('boxShadow');
      expect(style.outline).toBeUndefined();
    }
  });
});

describe('expansionMargin', () => {
  it('adds expansion gap for neighbor of selected card', () => {
    const baseOverlap = -10;
    expect(expansionMargin(true, baseOverlap)).toBe(baseOverlap + EXPANSION_GAP_PX);
  });

  it('returns base overlap for non-neighbor', () => {
    expect(expansionMargin(false, -10)).toBe(-10);
  });

  it('works with positive gap values', () => {
    expect(expansionMargin(true, 2)).toBe(2 + EXPANSION_GAP_PX);
    expect(expansionMargin(false, 2)).toBe(2);
  });

  it('works with zero overlap', () => {
    expect(expansionMargin(true, 0)).toBe(EXPANSION_GAP_PX);
    expect(expansionMargin(false, 0)).toBe(0);
  });
});

describe('smartHighlightStyle', () => {
  it('returns highlighted styles when true', () => {
    const style = smartHighlightStyle(true);
    expect(style.border).toBe('3px solid var(--color-game-status-active)');
    expect(style.boxShadow).toBe('0 0 10px rgba(92, 184, 92, 0.4), 0 0 20px rgba(92, 184, 92, 0.15)');
    expect(style.transition).toBe('border 0.15s, box-shadow 0.15s');
  });

  it('returns non-highlighted styles when false', () => {
    const style = smartHighlightStyle(false);
    expect(style.border).toBe('3px solid transparent');
    expect(style.boxShadow).toBe('none');
    expect(style.transition).toBe('border 0.15s, box-shadow 0.15s');
  });
});
