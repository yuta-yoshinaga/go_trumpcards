import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, OasisPokerResponse } from '../../../types/card';
import { formatOasispokerState } from './oasispokerFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard = { design: '' as CardDesign, value: 0 };

const betPhaseState: OasisPokerResponse = {
  playerHand: [],
  dealerHand: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  jackpotBet: 0,
  exchangeCount: 0,
  exchangeFee: 0,
  playBet: 0,
  result: 0,
  antePayout: 0,
  playPayout: 0,
  jackpotPayout: 0,
  totalPayout: 0,
  dealerQualified: false,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const exchangePhaseState: OasisPokerResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [card('SPADE', 10), card('HEART', 11), card('DIAMOND', 13), card('CLOVER', 5), card('SPADE', 7)],
  dealerHand: [card('HEART', 13), maskedCard, maskedCard, maskedCard, maskedCard],
  anteBet: 100,
  chips: 900,
};

const actionPhaseState: OasisPokerResponse = {
  ...exchangePhaseState,
  phase: 3,
  exchangeCount: 2,
  exchangeFee: 200,
  chips: 700,
};

const endPhasePlayerWins: OasisPokerResponse = {
  playerHand: [card('SPADE', 7), card('CLOVER', 7), card('HEART', 7), card('DIAMOND', 4), card('SPADE', 2)],
  dealerHand: [card('CLOVER', 5), card('DIAMOND', 5), card('HEART', 8), card('SPADE', 11), card('DIAMOND', 1)],
  phase: 4,
  chips: 1500,
  anteBet: 100,
  jackpotBet: 0,
  exchangeCount: 1,
  exchangeFee: 100,
  playBet: 200,
  result: 1,
  antePayout: 200,
  playPayout: 800,
  jackpotPayout: 0,
  totalPayout: 1000,
  dealerQualified: true,
  playerHandRank: 3,
  dealerHandRank: 1,
  message: 'Player wins!',
};

const endPhaseFold: OasisPokerResponse = {
  ...endPhasePlayerWins,
  result: -1,
  playBet: 0,
  antePayout: 0,
  playPayout: 0,
  totalPayout: 0,
  dealerHand: [],
  dealerQualified: false,
  message: 'Player folded.',
};

const endPhaseWithJackpot: OasisPokerResponse = {
  ...endPhasePlayerWins,
  jackpotBet: 10,
  jackpotPayout: 1000,
  totalPayout: 2000,
};

describe('formatOasispokerState', () => {
  it('formats bet phase with chips and phase name', () => {
    const result = formatOasispokerState(betPhaseState);
    expect(result).toContain('chips: 1000');
    expect(result).toContain('phase: BET');
    expect(result).not.toContain('Your hand');
    expect(result).not.toContain('Dealer');
    expect(result).not.toContain('ante:');
    expect(result).not.toContain('exchanged:');
  });

  it('formats exchange phase with player hand and masked dealer', () => {
    const result = formatOasispokerState(exchangePhaseState);
    expect(result).toContain('phase: EXCHANGE');
    expect(result).toContain('Your hand:');
    // First dealer card visible
    expect(result).toContain('♥K');
    // Remaining dealer cards masked
    expect(result).toContain('??');
    expect(result).toContain('ante: 100');
    // No exchange line yet (count is 0)
    expect(result).not.toContain('exchanged:');
  });

  it('formats action phase with exchange fee line', () => {
    const result = formatOasispokerState(actionPhaseState);
    expect(result).toContain('phase: ACTION');
    expect(result).toContain('exchanged: 2 (fee: 200)');
    // Dealer still masked in action phase
    expect(result).toContain('??');
  });

  it('formats end phase with full dealer hand and qualification', () => {
    const result = formatOasispokerState(endPhasePlayerWins);
    expect(result).toContain('phase: END');
    expect(result).toContain('Dealer:');
    expect(result).toContain('Dealer qualified: yes');
    expect(result).toContain('payout: ante=200 play=800 jackpot=0');
    expect(result).toContain('total: 1000');
    expect(result).toContain('Player wins!');
    // Exchange fee carried through to end
    expect(result).toContain('exchanged: 1 (fee: 100)');
  });

  it('formats end phase with dealer not qualified', () => {
    const state = { ...endPhasePlayerWins, dealerQualified: false };
    const result = formatOasispokerState(state);
    expect(result).toContain('Dealer qualified: no');
  });

  it('formats end phase after fold (no dealer hand)', () => {
    const result = formatOasispokerState(endPhaseFold);
    expect(result).toContain('phase: END');
    expect(result).not.toContain('Dealer:');
    expect(result).toContain('Player folded.');
    // playBet=0 → no play bet line
    expect(result).not.toContain('play bet:');
  });

  it('formats jackpot bet and payout', () => {
    const result = formatOasispokerState(endPhaseWithJackpot);
    expect(result).toContain('jackpot: 10');
    expect(result).toContain('jackpot=1000');
    expect(result).toContain('total: 2000');
  });

  it('formats unknown phase gracefully', () => {
    const state = { ...betPhaseState, phase: 99 };
    const result = formatOasispokerState(state);
    expect(result).toContain('phase: UNKNOWN');
  });

  it('omits message when empty', () => {
    const state = { ...betPhaseState, message: '' };
    const result = formatOasispokerState(state);
    expect(result.split('\n').filter((l) => l.trim())).not.toContain('');
  });

  it('includes message when present', () => {
    const state = { ...betPhaseState, message: 'Custom message' };
    const result = formatOasispokerState(state);
    expect(result).toContain('Custom message');
  });

  it('includes play bet when set', () => {
    const result = formatOasispokerState(endPhasePlayerWins);
    expect(result).toContain('play bet: 200');
  });
});
