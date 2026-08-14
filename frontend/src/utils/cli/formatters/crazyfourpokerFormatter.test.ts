import { describe, expect, it } from 'vitest';
import type { Card, CrazyFourPokerResponse } from '../../../types/card';
import { CRAZY_FOUR_POKER_RESULT } from '../../../types/games/crazyfourpoker';
import { CrazyFourPokerPhase } from '../../../types/phases';
import { formatCrazyFourPokerState } from './crazyfourpokerFormatter';

const ace: Card = { design: 'SPADE', value: 1 };
const king: Card = { design: 'HEART', value: 13 };

const base = {
  phase: CrazyFourPokerPhase.BET,
  playerHand: [],
  dealerHand: [],
  playerBest: [],
  dealerBest: [],
  playerHandRank: 0,
  dealerHandRank: 0,
  hasAcesOrBetter: false,
  maxMultiplier: 1,
  dealerQualifies: false,
  anteBet: 0,
  superBet: 0,
  queensUpBet: 0,
  playBet: 0,
  playMultiplier: 0,
  result: 0,
  payout: 0,
  chips: 1000,
  minTotalWager: 30,
  roundNumber: 1,
  remainingCards: 52,
  gameEndFlag: false,
  message: '',
} as unknown as CrazyFourPokerResponse;

const at = (over: Partial<CrazyFourPokerResponse>) => ({ ...base, ...over }) as CrazyFourPokerResponse;

describe('formatCrazyFourPokerState', () => {
  it('フェーズとチップを出す', () => {
    const out = formatCrazyFourPokerState(base);
    expect(out).toContain('BET');
    expect(out).toContain('chips: 1000');
  });

  it('賭ける前は賭けの行を出さない', () => {
    expect(formatCrazyFourPokerState(base)).not.toContain('Ante');
  });

  it('賭けた後は3本すべて出す', () => {
    const out = formatCrazyFourPokerState(at({ anteBet: 50, superBet: 50, queensUpBet: 20 }));
    expect(out).toContain('Ante 50');
    expect(out).toContain('Super Bonus 50');
    expect(out).toContain('Queens Up 20');
  });

  it('判断中は置ける倍率を出す', () => {
    const out = formatCrazyFourPokerState(
      at({
        phase: CrazyFourPokerPhase.DECIDE,
        anteBet: 50,
        superBet: 50,
        playerHand: [ace, ace, king, king, king],
        playerBest: [ace, ace, king, king],
        playerHandRank: 3,
        maxMultiplier: 3,
      }),
    );
    expect(out).toContain('multipliers available: 1-3');
    expect(out).toContain('Two Pair');
  });

  // **決着まではディーラーの手を出さない。**
  it('判断中はディーラーの手を出さない', () => {
    const out = formatCrazyFourPokerState(
      at({
        phase: CrazyFourPokerPhase.DECIDE,
        anteBet: 50,
        superBet: 50,
        playerHand: [ace, ace, king, king, king],
        playerBest: [ace, ace, king, king],
        playerHandRank: 3,
      }),
    );
    expect(out).toContain('You:');
    expect(out).not.toContain('Dealer:');
  });

  it('決着後はディーラーの手と不成立を出す', () => {
    const out = formatCrazyFourPokerState(
      at({
        phase: CrazyFourPokerPhase.RESULT,
        anteBet: 50,
        superBet: 50,
        playBet: 50,
        playerHand: [ace, ace, king, king, king],
        playerBest: [ace, ace, king, king],
        playerHandRank: 3,
        dealerHand: [king, king, ace, ace, ace],
        dealerBest: [king, king, ace, ace],
        dealerHandRank: 3,
        dealerQualifies: false,
        result: CRAZY_FOUR_POKER_RESULT.dealerNotQualified,
        payout: 200,
      }),
    );
    expect(out).toContain('Dealer:');
    expect(out).toContain('does not qualify');
    expect(out).toContain('dealer does not qualify');
    expect(out).toContain('net 50'); // 200 returned against 150 staked
  });

  it('決着の種類を出す', () => {
    for (const [key, want] of [
      [CRAZY_FOUR_POKER_RESULT.fold, 'folded'],
      [CRAZY_FOUR_POKER_RESULT.win, 'you win'],
      [CRAZY_FOUR_POKER_RESULT.lose, 'dealer wins'],
      [CRAZY_FOUR_POKER_RESULT.push, 'a push'],
    ] as const) {
      expect(formatCrazyFourPokerState(at({ result: key }))).toContain(want);
    }
  });

  it('資金切れを出す', () => {
    expect(formatCrazyFourPokerState(at({ gameEndFlag: true }))).toContain('Out of chips.');
  });
});
