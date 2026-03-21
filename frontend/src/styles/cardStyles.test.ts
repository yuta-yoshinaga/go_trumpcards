import { describe, expect, it } from 'bun:test';
import { playableCardStyle, selectedCardStyle, smartHighlightStyle } from './cardStyles';

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
