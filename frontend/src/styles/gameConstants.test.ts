import { describe, expect, it } from 'vitest';
import { PokerAction } from '../types/phases';
import { activeTurnClass, bettingActionName, finishedPlayerClass, handNameBadgeClass } from './gameConstants';

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

describe('handNameBadgeClass', () => {
  it('includes background and text color classes', () => {
    expect(handNameBadgeClass).toContain('bg-game-status-waiting');
    expect(handNameBadgeClass).toContain('text-game-text-strong');
  });
});

describe('activeTurnClass', () => {
  it('includes border and shadow classes', () => {
    expect(activeTurnClass).toContain('border-2');
    expect(activeTurnClass).toContain('border-game-status-waiting');
    expect(activeTurnClass).toContain('shadow-');
  });
});

describe('finishedPlayerClass', () => {
  it('is opacity-50', () => {
    expect(finishedPlayerClass).toBe('opacity-50');
  });
});
