import { describe, expect, it } from 'vitest';
import type { Card, FourCardPokerResponse } from '../../../types/card';
import { formatFourCardPokerState } from './fourcardpokerFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const baseState: FourCardPokerResponse = {
  message: '',
  playerHand: [card('SPADE', 5), card('HEART', 5), card('DIAMOND', 9), card('CLOVER', 11), card('SPADE', 2)],
  dealerHand: [card('HEART', 13)],
  playerBest: [],
  dealerBest: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  acesUpBet: 0,
  playBet: 0,
  playMultiplier: 0,
  result: 0,
  antePayout: 0,
  playPayout: 0,
  anteBonusPayout: 0,
  acesUpPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
};

describe('formatFourCardPokerState', () => {
  it('returns a loading placeholder for null state', () => {
    expect(formatFourCardPokerState(null)).toBe('Loading...');
  });

  it('prompts for a bet in the BET phase', () => {
    const out = formatFourCardPokerState(baseState);
    expect(out).toContain('Four Card Poker');
    expect(out).toContain('chips: 1000');
    expect(out).toContain('phase: BET');
    expect(out).toContain('Place a bet');
  });

  it('shows the hand and dealer upcard in the ACTION phase', () => {
    const out = formatFourCardPokerState({ ...baseState, phase: 2, anteBet: 100 });
    expect(out).toContain('phase: ACTION');
    expect(out).toContain('your hand: ♠5, ♥5');
    expect(out).toContain('dealer up: ♥K');
    expect(out).toContain('Make your play bet');
  });

  it('reveals best hands and payouts in the END phase', () => {
    const out = formatFourCardPokerState({
      ...baseState,
      phase: 3,
      anteBet: 100,
      playBet: 100,
      dealerHand: [card('HEART', 13), card('SPADE', 7), card('CLOVER', 3), card('DIAMOND', 4)],
      playerBest: [card('SPADE', 5), card('HEART', 5), card('CLOVER', 11), card('DIAMOND', 9)],
      dealerBest: [card('HEART', 13), card('SPADE', 7), card('CLOVER', 3), card('DIAMOND', 4)],
      antePayout: 100,
      anteBonusPayout: 0,
      playPayout: 100,
      acesUpPayout: 0,
      totalPayout: 200,
    });
    expect(out).toContain('phase: END');
    expect(out).toContain('dealer hand:');
    expect(out).toContain('your best:');
    expect(out).toContain('dealer best:');
    expect(out).toContain('total payout: 200');
  });

  it('appends a server message when present', () => {
    const out = formatFourCardPokerState({ ...baseState, message: 'Place your ante' });
    expect(out).toContain('Place your ante');
  });
});
