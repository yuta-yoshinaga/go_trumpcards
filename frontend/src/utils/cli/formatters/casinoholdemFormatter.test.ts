import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, CasinoHoldemResponse } from '../../../types/card';
import { CasinoHoldemPhase } from '../../../types/phases';
import { formatCasinoholdemState } from './casinoholdemFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard = { design: '' as CardDesign, value: 0 };

const baseState: CasinoHoldemResponse = {
  playerHand: [],
  dealerHand: [],
  community: [],
  phase: CasinoHoldemPhase.BET,
  chips: 1000,
  anteBet: 0,
  bonusBet: 0,
  callBet: 0,
  result: 0,
  dealerQualify: false,
  antePayout: 0,
  callPayout: 0,
  bonusPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const flopState: CasinoHoldemResponse = {
  ...baseState,
  phase: CasinoHoldemPhase.FLOP,
  playerHand: [card('SPADE', 1), card('SPADE', 13)],
  dealerHand: [maskedCard, maskedCard],
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10)],
  anteBet: 100,
  bonusBet: 10,
  chips: 890,
  playerHandRank: 5,
};

const endStateCallWin: CasinoHoldemResponse = {
  ...baseState,
  phase: CasinoHoldemPhase.END,
  playerHand: [card('SPADE', 1), card('SPADE', 13)],
  dealerHand: [card('HEART', 7), card('DIAMOND', 5)],
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10), card('CLOVER', 2), card('HEART', 4)],
  anteBet: 100,
  bonusBet: 0,
  callBet: 200,
  result: 1,
  dealerQualify: true,
  antePayout: 100 + 100 * 100, // royal flush
  callPayout: 400,
  totalPayout: 100 + 100 * 100 + 400,
  playerHandRank: 9,
  dealerHandRank: 0,
  message: 'Player wins!',
};

const endStateNoQualify: CasinoHoldemResponse = {
  ...endStateCallWin,
  dealerQualify: false,
  callPayout: 200, // call pushes when dealer doesn't qualify
  totalPayout: 100 + 100 * 100 + 200,
};

const endStateFold: CasinoHoldemResponse = {
  ...baseState,
  phase: CasinoHoldemPhase.END,
  playerHand: [card('SPADE', 2), card('HEART', 5)],
  dealerHand: [maskedCard, maskedCard],
  community: [card('CLOVER', 3), card('HEART', 7), card('DIAMOND', 9)],
  anteBet: 100,
  callBet: 0,
  result: -1,
  dealerQualify: false,
  message: 'Player folded.',
};

describe('formatCasinoholdemState', () => {
  it('formats bet phase with no hands or board', () => {
    const out = formatCasinoholdemState(baseState);
    expect(out).toContain("Casino Hold'em");
    expect(out).toContain('chips: 1000');
    expect(out).toContain('phase: BET');
    expect(out).not.toContain('Your hand:');
    expect(out).not.toContain('Dealer:');
    expect(out).not.toContain('Board:');
  });

  it('formats flop with player hand, masked dealer, and visible board', () => {
    const out = formatCasinoholdemState(flopState);
    expect(out).toContain('phase: FLOP');
    expect(out).toContain('Your hand:');
    expect(out).toContain('Dealer:');
    expect(out).toMatch(/Dealer:\s+\?\?,\s+\?\?/);
    expect(out).toContain('Board:');
    expect(out).toContain('ante: 100');
    expect(out).toContain('AA bonus: 10');
    expect(out).not.toContain('call:');
  });

  it('reveals dealer hand and shows qualify + payout breakdown after Call win', () => {
    const out = formatCasinoholdemState(endStateCallWin);
    expect(out).toContain('phase: END');
    expect(out).toContain('Board:');
    expect(out).toContain('Dealer:');
    expect(out).not.toContain('??');
    expect(out).toContain('Dealer qualifies');
    expect(out).toContain('payout: ante=10100 call=400 bonus=0');
    expect(out).toContain('total: 10500');
    expect(out).toContain('Player wins!');
  });

  it('shows "Dealer does not qualify" branch after Call no-qualify', () => {
    const out = formatCasinoholdemState(endStateNoQualify);
    expect(out).toContain('Dealer does not qualify');
    expect(out).toContain('payout: ante=10100 call=200 bonus=0');
  });

  it('formats fold end state with masked dealer and no qualify line', () => {
    const out = formatCasinoholdemState(endStateFold);
    expect(out).toContain('phase: END');
    expect(out).toContain('Dealer:');
    expect(out).toMatch(/Dealer:\s+\?\?,\s+\?\?/);
    expect(out).toContain('Player folded.');
    expect(out).toContain('ante: 100');
    expect(out).not.toContain('call:');
    expect(out).not.toContain('Dealer qualifies');
    expect(out).not.toContain('Dealer does not qualify');
  });

  it('formats unknown phase gracefully', () => {
    const out = formatCasinoholdemState({ ...baseState, phase: 99 });
    expect(out).toContain('phase: UNKNOWN');
  });

  it('omits zero-value bet lines', () => {
    const out = formatCasinoholdemState({ ...baseState, anteBet: 0, bonusBet: 0, callBet: 0 });
    expect(out).not.toContain('ante:');
    expect(out).not.toContain('AA bonus:');
    expect(out).not.toContain('call:');
  });
});
