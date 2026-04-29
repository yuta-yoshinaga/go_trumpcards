import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, TexasHoldemBonusResponse } from '../../../types/card';
import { TexasHoldemBonusPhase } from '../../../types/phases';
import { formatTexasholdembonusState } from './texasholdembonusFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard = { design: '' as CardDesign, value: 0 };

const baseState: TexasHoldemBonusResponse = {
  playerHand: [],
  dealerHand: [],
  community: [],
  phase: TexasHoldemBonusPhase.BET,
  chips: 1000,
  anteBet: 0,
  bonusBet: 0,
  flopBet: 0,
  turnBet: 0,
  riverBet: 0,
  totalPlayBet: 0,
  result: 0,
  antePayout: 0,
  playPayout: 0,
  bonusPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const preFlopState: TexasHoldemBonusResponse = {
  ...baseState,
  phase: TexasHoldemBonusPhase.PRE_FLOP,
  playerHand: [card('SPADE', 12), card('CLOVER', 4)],
  dealerHand: [maskedCard, maskedCard],
  anteBet: 100,
  bonusBet: 10,
  flopBet: 200,
  totalPlayBet: 200,
  chips: 690,
};

const flopState: TexasHoldemBonusResponse = {
  ...preFlopState,
  phase: TexasHoldemBonusPhase.FLOP,
  community: [card('HEART', 11), card('CLOVER', 1), card('CLOVER', 12)],
  playerHandRank: 1,
};

const endStateWin: TexasHoldemBonusResponse = {
  ...baseState,
  phase: TexasHoldemBonusPhase.END,
  playerHand: [card('SPADE', 12), card('CLOVER', 4)],
  dealerHand: [card('HEART', 13), card('CLOVER', 5)],
  community: [card('HEART', 11), card('CLOVER', 1), card('CLOVER', 12), card('DIAMOND', 2), card('HEART', 2)],
  anteBet: 100,
  bonusBet: 10,
  flopBet: 200,
  turnBet: 100,
  riverBet: 100,
  totalPlayBet: 400,
  result: 1,
  antePayout: 200,
  playPayout: 800,
  bonusPayout: 0,
  totalPayout: 1000,
  chips: 1390,
  playerHandRank: 3,
  dealerHandRank: 1,
  message: 'Player wins!',
};

const endStateFold: TexasHoldemBonusResponse = {
  ...baseState,
  phase: TexasHoldemBonusPhase.END,
  playerHand: [card('SPADE', 2), card('HEART', 5)],
  anteBet: 100,
  result: -1,
  message: 'Player folded.',
};

describe('formatTexasholdembonusState', () => {
  it('formats bet phase header without hands', () => {
    const out = formatTexasholdembonusState(baseState);
    expect(out).toContain("Texas Hold'em Bonus Poker");
    expect(out).toContain('chips: 1000');
    expect(out).toContain('phase: BET');
    expect(out).not.toContain('Your hand:');
    expect(out).not.toContain('Dealer:');
    expect(out).not.toContain('Board:');
  });

  it('formats pre-flop with player hand and fully masked dealer', () => {
    const out = formatTexasholdembonusState(preFlopState);
    expect(out).toContain('phase: PRE-FLOP');
    expect(out).toContain('Your hand:');
    expect(out).toContain('Dealer:');
    expect(out).toMatch(/Dealer:\s+\?\?,\s+\?\?/);
    expect(out).not.toContain('Board:');
    expect(out).toContain('ante: 100');
    expect(out).toContain('bonus: 10');
    expect(out).toContain('play bets: 200');
  });

  it('formats flop with community cards', () => {
    const out = formatTexasholdembonusState(flopState);
    expect(out).toContain('phase: FLOP');
    expect(out).toContain('Board:');
    expect(out).toContain('Your hand:');
    expect(out).toContain('Dealer:');
  });

  it('reveals dealer hand and shows payout breakdown at showdown', () => {
    const out = formatTexasholdembonusState(endStateWin);
    expect(out).toContain('phase: END');
    expect(out).toContain('Board:');
    expect(out).toContain('Dealer:');
    expect(out).not.toContain('??');
    expect(out).toContain('payout: ante=200 play=800 bonus=0');
    expect(out).toContain('total: 1000');
    expect(out).toContain('Player wins!');
  });

  it('formats fold end state without dealer hand or payout line', () => {
    const out = formatTexasholdembonusState(endStateFold);
    expect(out).toContain('phase: END');
    expect(out).not.toContain('Dealer:');
    expect(out).toContain('Player folded.');
    expect(out).toContain('ante: 100');
    expect(out).not.toContain('play bets:');
  });

  it('formats unknown phase gracefully', () => {
    const out = formatTexasholdembonusState({ ...baseState, phase: 99 });
    expect(out).toContain('phase: UNKNOWN');
  });

  it('omits zero-value bet lines', () => {
    const out = formatTexasholdembonusState({ ...baseState, anteBet: 0, bonusBet: 0, totalPlayBet: 0 });
    expect(out).not.toContain('ante:');
    expect(out).not.toContain('bonus:');
    expect(out).not.toContain('play bets:');
  });
});
