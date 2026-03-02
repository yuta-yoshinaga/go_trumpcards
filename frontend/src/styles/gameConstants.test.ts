import { describe, expect, it } from 'vitest';
import { HoldemAction, PokerAction } from '../types/phases';
import {
  activeTurnStyle,
  finishedPlayerStyle,
  HOLDEM_ACTION_NAMES,
  handNameBadgeStyle,
  POKER_ACTION_NAMES,
} from './gameConstants';

describe('POKER_ACTION_NAMES', () => {
  it('maps all PokerAction values', () => {
    expect(POKER_ACTION_NAMES[PokerAction.FOLD]).toBe('フォールド');
    expect(POKER_ACTION_NAMES[PokerAction.CHECK]).toBe('チェック');
    expect(POKER_ACTION_NAMES[PokerAction.CALL]).toBe('コール');
    expect(POKER_ACTION_NAMES[PokerAction.BET]).toBe('ベット');
    expect(POKER_ACTION_NAMES[PokerAction.RAISE]).toBe('レイズ');
    expect(POKER_ACTION_NAMES[PokerAction.ALL_IN]).toBe('オールイン');
  });
});

describe('HOLDEM_ACTION_NAMES', () => {
  it('maps all HoldemAction values', () => {
    expect(HOLDEM_ACTION_NAMES[HoldemAction.FOLD]).toBe('フォールド');
    expect(HOLDEM_ACTION_NAMES[HoldemAction.CHECK]).toBe('チェック');
    expect(HOLDEM_ACTION_NAMES[HoldemAction.CALL]).toBe('コール');
    expect(HOLDEM_ACTION_NAMES[HoldemAction.BET]).toBe('ベット');
    expect(HOLDEM_ACTION_NAMES[HoldemAction.RAISE]).toBe('レイズ');
    expect(HOLDEM_ACTION_NAMES[HoldemAction.ALL_IN]).toBe('オールイン');
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
