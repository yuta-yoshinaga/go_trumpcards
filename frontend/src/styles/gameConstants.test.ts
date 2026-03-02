import { describe, expect, it } from 'vitest';
import { PokerAction } from '../types/phases';
import {
  activeTurnStyle,
  BETTING_ACTION_NAMES,
  finishedPlayerStyle,
  HOLDEM_ACTION_NAMES,
  handNameBadgeStyle,
  POKER_ACTION_NAMES,
} from './gameConstants';

describe('BETTING_ACTION_NAMES', () => {
  it('maps all action values', () => {
    expect(BETTING_ACTION_NAMES[PokerAction.FOLD]).toBe('フォールド');
    expect(BETTING_ACTION_NAMES[PokerAction.CHECK]).toBe('チェック');
    expect(BETTING_ACTION_NAMES[PokerAction.CALL]).toBe('コール');
    expect(BETTING_ACTION_NAMES[PokerAction.BET]).toBe('ベット');
    expect(BETTING_ACTION_NAMES[PokerAction.RAISE]).toBe('レイズ');
    expect(BETTING_ACTION_NAMES[PokerAction.ALL_IN]).toBe('オールイン');
  });

  it('POKER_ACTION_NAMES and HOLDEM_ACTION_NAMES are aliases', () => {
    expect(POKER_ACTION_NAMES).toBe(BETTING_ACTION_NAMES);
    expect(HOLDEM_ACTION_NAMES).toBe(BETTING_ACTION_NAMES);
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
