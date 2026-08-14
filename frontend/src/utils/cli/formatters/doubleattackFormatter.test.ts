import { describe, expect, it } from 'vitest';
import type { Card, DoubleAttackResponse } from '../../../types/card';
import { DOUBLE_ATTACK_RESULT } from '../../../types/games/doubleattack';
import { DoubleAttackPhase } from '../../../types/phases';
import { formatDoubleAttackState } from './doubleattackFormatter';

const card = (value: number): Card => ({ design: 'SPADE', value });

const hand = (over: Partial<DoubleAttackResponse['hands'][number]> = {}) =>
  ({
    cards: [card(13), card(7)],
    score: 17,
    bet: 100,
    isSoft: false,
    stood: false,
    doubled: false,
    busted: false,
    blackjack: false,
    result: 0,
    ...over,
  }) as DoubleAttackResponse['hands'][number];

const base = {
  phase: DoubleAttackPhase.BET,
  hands: [],
  activeHand: 0,
  dealerCards: [],
  dealerScore: 0,
  dealerHoleDealt: false,
  maxAttackBet: 0,
  canDouble: false,
  canSplit: false,
  anteBet: 0,
  attackBet: 0,
  bustItBet: 0,
  payout: 0,
  bustItPayout: 0,
  chips: 1000,
  roundNumber: 1,
  remainingCards: 384,
  gameEndFlag: false,
  message: '',
} as unknown as DoubleAttackResponse;

const at = (over: Partial<DoubleAttackResponse>) => ({ ...base, ...over }) as DoubleAttackResponse;

describe('formatDoubleAttackState', () => {
  it('フェーズとチップを出す', () => {
    const out = formatDoubleAttackState(base);
    expect(out).toContain('BET');
    expect(out).toContain('chips: 1000');
  });

  it('賭けた後は3本すべて出す', () => {
    const out = formatDoubleAttackState(at({ anteBet: 50, attackBet: 25, bustItBet: 20 }));
    expect(out).toContain('Ante 50');
    expect(out).toContain('Extra bet 25');
    expect(out).toContain('Bust It 20');
  });

  // **追加ベットの前は 1 枚だけで、点数は出さない。**
  it('アップカードだけのときは点数を出さない', () => {
    const out = formatDoubleAttackState(
      at({
        phase: DoubleAttackPhase.ATTACK,
        anteBet: 50,
        dealerCards: [card(6)],
        dealerHoleDealt: false,
        maxAttackBet: 50,
        hands: [hand()],
      }),
    );
    expect(out).toContain('second card comes after the extra bet');
    expect(out).toContain('Extra bet limit: 50');
    expect(out).not.toMatch(/Dealer:.*=/);
  });

  it('2枚目が配られたら点数を出す', () => {
    const out = formatDoubleAttackState(
      at({
        phase: DoubleAttackPhase.PLAY,
        anteBet: 50,
        dealerCards: [card(6), card(9)],
        dealerScore: 15,
        dealerHoleDealt: true,
        hands: [hand()],
      }),
    );
    expect(out).toContain('= 15');
    expect(out).not.toContain('second card comes after');
  });

  it('操作中の手札に印を付ける', () => {
    const out = formatDoubleAttackState(
      at({
        phase: DoubleAttackPhase.PLAY,
        anteBet: 50,
        hands: [hand(), hand()],
        activeHand: 1,
        dealerCards: [card(6), card(9)],
        dealerHoleDealt: true,
      }),
    );
    expect(out).toContain('*[2]');
    expect(out).toContain(' [1]');
  });

  it('手札の状態を出す', () => {
    const out = formatDoubleAttackState(
      at({ anteBet: 50, hands: [hand({ busted: true, doubled: true, blackjack: true })] }),
    );
    expect(out).toContain('bust');
    expect(out).toContain('doubled');
    expect(out).toContain('blackjack');
  });

  it('決着では手札ごとの結果と Bust It を出す', () => {
    const out = formatDoubleAttackState(
      at({
        phase: DoubleAttackPhase.RESULT,
        anteBet: 50,
        hands: [hand({ result: DOUBLE_ATTACK_RESULT.blackjack })],
        bustItPayout: 150,
        dealerCards: [card(6), card(9)],
        dealerHoleDealt: true,
      }),
    );
    expect(out).toContain('blackjack (1:1)');
    expect(out).toContain('Bust It pays: 150');
  });

  it('資金切れを出す', () => {
    expect(formatDoubleAttackState(at({ gameEndFlag: true }))).toContain('Out of chips.');
  });
});
