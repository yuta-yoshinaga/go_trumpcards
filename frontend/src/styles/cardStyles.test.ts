import { describe, expect, it } from 'vitest';
import {
  expansionMargin,
  focusRingCard,
  highlightCardStyle,
  meldCardStyle,
  playableCardStyle,
  playableRingStyle,
  selectedCardStyle,
  smartHighlightStyle,
  trumpRingStyle,
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
  // The indicator must survive the inline styles every card button sets.
  //
  // Card buttons write inline styles on BOTH channels a Tailwind utility could
  // use: `boxShadow` (selectedCardStyle & friends, whose "off" branch sets
  // `'none'`) and `outline` (trumpRingStyle / meldCardStyle / playableRingStyle,
  // where meldCardStyle applies to *every* hand card during the GinRummy and
  // Chinchon discard phases). Inline styles win, so the indicator lives in a
  // stylesheet rule with `!important` instead. See issue #5359.
  it('delegates to the stylesheet rule rather than a utility inline styles can erase', () => {
    expect(focusRingCard).toContain('card-focus-ring');
    expect(focusRingCard).toContain('rounded-lg');
  });

  it('uses no ring or outline utility, either of which an inline style would beat', () => {
    // Anchored to a class boundary: `card-focus-ring` legitimately ends in
    // "ring", and a substring match would reject the very class we want.
    const utility = (name: string) => new RegExp(`(^|[\\s:])${name}(-|$|\\s)`);
    expect(focusRingCard).not.toMatch(utility('ring'));
    expect(focusRingCard).not.toMatch(utility('outline'));
  });

  it('does not disable the browser default outline', () => {
    expect(focusRingCard).not.toContain('outline-none');
  });
});

describe('the inline styles the focus indicator has to coexist with', () => {
  // Enumerated deliberately. The first fix for #5359 moved the focus ring from
  // `boxShadow` to `outline` and reintroduced the same collision, because the
  // interaction test only covered the boxShadow-setting helpers. Listing both
  // families means a new decorative helper shows up here as a failing case
  // rather than as an invisible focus ring.
  it('boxShadow family always sets boxShadow, in both branches', () => {
    for (const style of [
      selectedCardStyle(true),
      selectedCardStyle(false),
      playableCardStyle(true),
      playableCardStyle(false),
      highlightCardStyle(),
      smartHighlightStyle(true),
      smartHighlightStyle(false),
    ]) {
      expect(style).toHaveProperty('boxShadow');
    }
  });

  it('outline family always sets outline, in both branches', () => {
    for (const style of [trumpRingStyle(), meldCardStyle(true), meldCardStyle(false), playableRingStyle()]) {
      expect(style).toHaveProperty('outline');
      expect(style.outline).not.toBe('none');
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
