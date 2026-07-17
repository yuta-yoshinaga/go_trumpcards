import { describe, expect, it } from 'vitest';
import type { Card, ChinesePokerResponse } from '../../../types/card';
import { formatChinesePokerState } from './chinesepokerFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const baseState: ChinesePokerResponse = {
  message: '',
  playerCards: [card('SPADE', 2), card('HEART', 5), card('DIAMOND', 9)],
  dealerCards: [],
  playerFront: [],
  playerMiddle: [],
  playerBack: [],
  dealerFront: [],
  dealerMiddle: [],
  dealerBack: [],
  phase: 1,
  chips: 1000,
  bet: 0,
  result: 0,
  frontResult: 0,
  middleResult: 0,
  backResult: 0,
  payout: 0,
  playerFrontRank: 0,
  playerMiddleRank: 0,
  playerBackRank: 0,
  dealerFrontRank: 0,
  dealerMiddleRank: 0,
  dealerBackRank: 0,
  playerRoyalty: 0,
  dealerRoyalty: 0,
  scoop: false,
};

describe('formatChinesePokerState', () => {
  it('returns a loading placeholder for null state', () => {
    expect(formatChinesePokerState(null)).toBe('Loading...');
  });

  it('prompts for a bet in the BET phase', () => {
    const out = formatChinesePokerState(baseState);
    expect(out).toContain('Chinese Poker');
    expect(out).toContain('chips: 1000');
    expect(out).toContain('phase: BET');
    expect(out).toContain('Place a bet');
  });

  it('shows the indexed hand in the SET HANDS phase', () => {
    const out = formatChinesePokerState({ ...baseState, phase: 2, bet: 50 });
    expect(out).toContain('phase: SET HANDS');
    expect(out).toContain('[0]♠2');
    expect(out).toContain('[2]♦9');
    expect(out).toContain('Set with:');
  });

  it('renders both hands and per-row results at END', () => {
    const out = formatChinesePokerState({
      ...baseState,
      phase: 3,
      bet: 50,
      playerFront: [card('SPADE', 2), card('SPADE', 3), card('SPADE', 4)],
      playerMiddle: [card('HEART', 5)],
      playerBack: [card('CLOVER', 10)],
      dealerFront: [card('DIAMOND', 6)],
      frontResult: 1,
      middleResult: -1,
      backResult: 0,
      payout: 100,
    });
    expect(out).toContain('phase: END');
    expect(out).toContain('front:  ♠2, ♠3, ♠4  [WIN]');
    expect(out).toContain('[LOSE]');
    expect(out).toContain('[PUSH]');
    expect(out).toContain('dealer front:  ♦6');
    expect(out).toContain('payout: 100');
  });

  it('announces a scoop and royalty bonus at END', () => {
    const out = formatChinesePokerState({
      ...baseState,
      phase: 3,
      scoop: true,
      playerRoyalty: 6,
    });
    expect(out).toContain('SCOOP!');
    expect(out).toContain('royalty bonus: +6');
  });

  it('appends a server message when present', () => {
    const out = formatChinesePokerState({ ...baseState, message: 'Arrange your hand' });
    expect(out).toContain('Arrange your hand');
  });
});
