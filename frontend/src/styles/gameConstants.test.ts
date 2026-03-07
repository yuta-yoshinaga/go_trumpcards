import { describe, expect, it } from 'vitest';
import { PokerAction } from '../types/phases';
import { activeTurnStyle, bettingActionName, finishedPlayerStyle, handNameBadgeStyle } from './gameConstants';

describe('bettingActionName', () => {
  it('maps all action values', () => {
    expect(bettingActionName(PokerAction.FOLD)).toBe('フォールド');
    expect(bettingActionName(PokerAction.CHECK)).toBe('チェック');
    expect(bettingActionName(PokerAction.CALL)).toBe('コール');
    expect(bettingActionName(PokerAction.BET)).toBe('ベット');
    expect(bettingActionName(PokerAction.RAISE)).toBe('レイズ');
    expect(bettingActionName(PokerAction.ALL_IN)).toBe('オールイン');
  });

  it('returns 不明 for unknown action', () => {
    expect(bettingActionName(99)).toBe('不明');
  });
});

describe('handNameBadgeStyle', () => {
  it('has correct background and color', () => {
    expect(handNameBadgeStyle.background).toBe('#f0ad4e');
    expect(handNameBadgeStyle.color).toBe('#222');
  });
});

describe('activeTurnStyle', () => {
  it('has correct border and boxShadow', () => {
    expect(activeTurnStyle.border).toBe('2px solid #f0ad4e');
    expect(activeTurnStyle.boxShadow).toBe('0 0 12px #f0ad4e');
  });
});

describe('finishedPlayerStyle', () => {
  it('has opacity 0.5', () => {
    expect(finishedPlayerStyle.opacity).toBe(0.5);
  });
});
