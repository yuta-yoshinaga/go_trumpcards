import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, HighCardFlushResponse } from '../../../types/card';
import { HighCardFlushPhase } from '../../../types/phases';
import { formatHighcardflushState } from './highcardflushFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

const baseState: HighCardFlushResponse = {
  playerHand: [],
  dealerHand: [],
  phase: HighCardFlushPhase.BET,
  chips: 1000,
  anteBet: 0,
  flushBonusBet: 0,
  straightFlushBet: 0,
  raiseBet: 0,
  result: 0,
  antePayout: 0,
  raisePayout: 0,
  flushBonusPayout: 0,
  straightFlushPayout: 0,
  totalPayout: 0,
  dealerQualified: false,
  playerFlushLen: 0,
  dealerFlushLen: 0,
  playerStraightFlushLen: 0,
  maxRaiseMultiplier: 3,
  message: '',
};

const actionState: HighCardFlushResponse = {
  ...baseState,
  phase: HighCardFlushPhase.ACTION,
  playerHand: [card('SPADE', 12), card('SPADE', 9), card('SPADE', 4), card('HEART', 7), card('CLOVER', 2)],
  dealerHand: [card('DIAMOND', 13), card('DIAMOND', 8), card('DIAMOND', 3), card('HEART', 5), card('CLOVER', 9)],
  anteBet: 100,
  flushBonusBet: 10,
  playerFlushLen: 3,
  chips: 890,
};

const endState: HighCardFlushResponse = {
  ...actionState,
  phase: HighCardFlushPhase.END,
  raiseBet: 200,
  dealerFlushLen: 3,
  dealerQualified: true,
  result: 1,
  antePayout: 100,
  raisePayout: 200,
  flushBonusPayout: 20,
  totalPayout: 320,
};

describe('formatHighcardflushState', () => {
  it('renders the header, chips, and phase', () => {
    const out = formatHighcardflushState(baseState);
    expect(out).toContain('High Card Flush');
    expect(out).toContain('chips: 1000');
    expect(out).toContain('phase: BET');
  });

  it('hides the dealer hand during the action phase', () => {
    const out = formatHighcardflushState(actionState);
    expect(out).toContain('Dealer: (hidden)');
    expect(out).toContain('Your longest flush: 3');
    expect(out).not.toContain('♦');
  });

  it('reveals the dealer hand and payouts at end', () => {
    const out = formatHighcardflushState(endState);
    expect(out).toContain('Dealer:');
    expect(out).toContain('Dealer longest flush: 3');
    expect(out).toContain('raise=200');
    expect(out).toContain('total: 320');
  });

  it('marks an unqualified dealer at end', () => {
    const out = formatHighcardflushState({ ...endState, dealerQualified: false });
    expect(out).toContain('not qualified');
  });

  it('renders UNKNOWN for an unexpected phase', () => {
    const out = formatHighcardflushState({ ...baseState, phase: 99 });
    expect(out).toContain('phase: UNKNOWN');
  });
});
