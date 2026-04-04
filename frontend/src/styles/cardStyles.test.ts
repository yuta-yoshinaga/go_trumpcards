import { describe, expect, it } from 'vitest';
import {
  expansionMargin,
  focusRingCard,
  playableCardStyle,
  selectedCardStyle,
  smartHighlightStyle,
} from './cardStyles';

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
  it('includes focus-visible ring classes', () => {
    expect(focusRingCard).toContain('focus-visible:outline-none');
    expect(focusRingCard).toContain('focus-visible:ring-2');
    expect(focusRingCard).toContain('focus-visible:ring-ds-accent');
    expect(focusRingCard).toContain('rounded-lg');
  });
});

describe('expansionMargin', () => {
  it('adds expansion gap for neighbor of selected card', () => {
    const baseOverlap = -10;
    expect(expansionMargin(true, baseOverlap)).toBe(-10 + 12);
  });

  it('returns base overlap for non-neighbor', () => {
    expect(expansionMargin(false, -10)).toBe(-10);
  });

  it('works with positive gap values', () => {
    expect(expansionMargin(true, 2)).toBe(14);
    expect(expansionMargin(false, 2)).toBe(2);
  });

  it('works with zero overlap', () => {
    expect(expansionMargin(true, 0)).toBe(12);
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
