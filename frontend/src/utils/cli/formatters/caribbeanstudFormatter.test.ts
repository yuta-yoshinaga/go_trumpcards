import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, CaribbeanStudResponse } from '../../../types/card';
import { formatCaribbeanstudState } from './caribbeanstudFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard = { design: '' as CardDesign, value: 0 };

const betPhaseState: CaribbeanStudResponse = {
  playerHand: [],
  dealerHand: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  jackpotBet: 0,
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

const actionPhaseState: CaribbeanStudResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [card('SPADE', 10), card('HEART', 11), card('DIAMOND', 13), card('CLOVER', 5), card('SPADE', 7)],
  dealerHand: [card('HEART', 13), maskedCard, maskedCard, maskedCard, maskedCard],
  anteBet: 100,
  chips: 900,
};

const endPhasePlayerWins: CaribbeanStudResponse = {
  playerHand: [card('SPADE', 7), card('CLOVER', 7), card('HEART', 7), card('DIAMOND', 4), card('SPADE', 2)],
  dealerHand: [card('CLOVER', 5), card('DIAMOND', 5), card('HEART', 8), card('SPADE', 11), card('DIAMOND', 1)],
  phase: 3,
  chips: 1500,
  anteBet: 100,
  jackpotBet: 0,
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

const endPhaseFold: CaribbeanStudResponse = {
  ...endPhasePlayerWins,
  result: -1,
  playBet: 0,
  antePayout: 0,
  playPayout: 0,
  totalPayout: 0,
  dealerHand: [],
  message: 'Player folded.',
};

const endPhaseWithJackpot: CaribbeanStudResponse = {
  ...endPhasePlayerWins,
  jackpotBet: 10,
  jackpotPayout: 1000,
  totalPayout: 2000,
};

describe('formatCaribbeanstudState', () => {
  it('formats bet phase with chips and phase name', () => {
    const result = formatCaribbeanstudState(betPhaseState);
    expect(result).toContain('chips: 1000');
    expect(result).toContain('phase: BET');
    expect(result).not.toContain('Your hand');
    expect(result).not.toContain('Dealer');
  });

  it('formats action phase with player hand and masked dealer hand', () => {
    const result = formatCaribbeanstudState(actionPhaseState);
    expect(result).toContain('phase: ACTION');
    expect(result).toContain('Your hand:');
    // First dealer card visible
    expect(result).toContain('♥K');
    // Remaining dealer cards hidden
    expect(result).toContain('??');
    expect(result).toContain('ante: 100');
  });

  it('formats end phase with full dealer hand and qualification', () => {
    const result = formatCaribbeanstudState(endPhasePlayerWins);
    expect(result).toContain('phase: END');
    expect(result).toContain('Dealer:');
    expect(result).toContain('Dealer qualified: yes');
    expect(result).toContain('payout: ante=200 play=800 jackpot=0');
    expect(result).toContain('total: 1000');
    expect(result).toContain('Player wins!');
  });

  it('formats end phase after fold (no dealer hand)', () => {
    const result = formatCaribbeanstudState(endPhaseFold);
    expect(result).toContain('phase: END');
    expect(result).not.toContain('Dealer:');
    expect(result).toContain('Player folded.');
  });

  it('formats jackpot bets and payout', () => {
    const result = formatCaribbeanstudState(endPhaseWithJackpot);
    expect(result).toContain('jackpot: 10');
    expect(result).toContain('jackpot=1000');
    expect(result).toContain('total: 2000');
  });

  it('formats unknown phase gracefully', () => {
    const state = { ...betPhaseState, phase: 99 };
    const result = formatCaribbeanstudState(state);
    expect(result).toContain('phase: UNKNOWN');
  });

  it('includes play bet when set', () => {
    const state = { ...endPhasePlayerWins };
    const result = formatCaribbeanstudState(state);
    expect(result).toContain('play bet: 200');
  });
});
