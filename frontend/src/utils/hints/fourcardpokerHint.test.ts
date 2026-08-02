import { describe, expect, it } from 'vitest';
import type { Card, FourCardPokerResponse } from '../../types/card';
import { FourCardPokerPhase } from '../../types/phases';
import { bestRank, FOUR_CARD_RANK, getFourCardPokerHint } from './fourcardpokerHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function base(overrides: Partial<FourCardPokerResponse> = {}): FourCardPokerResponse {
  return {
    playerHand: [card('SPADE', 5), card('HEART', 9), card('CLOVER', 12), card('DIAMOND', 3), card('SPADE', 7)],
    dealerHand: [],
    playerBest: [],
    dealerBest: [],
    phase: FourCardPokerPhase.ACTION,
    chips: 1000,
    anteBet: 10,
    acesUpBet: 0,
    playBet: 0,
    playMultiplier: 0,
    result: 0,
    antePayout: 0,
    playPayout: 0,
    anteBonusPayout: 0,
    acesUpPayout: 0,
    totalPayout: 0,
    // **常に 0 のまま。**`updateBestHands` は fold と resolve でしか走らない。
    playerHandRank: 0,
    dealerHandRank: 0,
    message: '',
    ...overrides,
  } as FourCardPokerResponse;
}

describe('getFourCardPokerHint', () => {
  it('suggests betting while chips remain', () => {
    expect(getFourCardPokerHint(base({ phase: FourCardPokerPhase.BET }))?.targetAction).toBe('bet');
  });

  it('says nothing in the bet phase once the chips are gone', () => {
    expect(getFourCardPokerHint(base({ phase: FourCardPokerPhase.BET, chips: 0 }))).toBeNull();
  });

  it('returns null after the showdown', () => {
    expect(getFourCardPokerHint(base({ phase: FourCardPokerPhase.END }))).toBeNull();
  });

  it('folds a hand that makes no more than a pair', () => {
    expect(getFourCardPokerHint(base())?.targetAction).toBe('fold');
  });

  it('pushes the maximum on three of a kind', () => {
    const s = base({
      playerHand: [card('SPADE', 8), card('HEART', 8), card('CLOVER', 8), card('DIAMOND', 3), card('SPADE', 2)],
    });
    expect(getFourCardPokerHint(s)?.targetAction).toBe('play-3');
  });

  it('bets the minimum on two pair', () => {
    const s = base({
      playerHand: [card('SPADE', 8), card('HEART', 8), card('CLOVER', 4), card('DIAMOND', 4), card('SPADE', 2)],
    });
    expect(getFourCardPokerHint(s)?.targetAction).toBe('play-1');
  });

  it('ranks a flush above a straight, unlike five-card poker', () => {
    // **順位そのものを見る。**どちらも行動は play-1 なので、戻り値だけを見ると
    // 順序を入れ替えても気づけない（その負のコントロールが実際に素通りした）。
    const flushHand = [card('SPADE', 2), card('SPADE', 5), card('SPADE', 9), card('SPADE', 12)];
    const straightHand = [card('SPADE', 5), card('HEART', 6), card('CLOVER', 7), card('DIAMOND', 8)];
    expect(bestRank(flushHand)).toBe(FOUR_CARD_RANK.FLUSH);
    expect(bestRank(straightHand)).toBe(FOUR_CARD_RANK.STRAIGHT);
    expect(bestRank(flushHand)).toBeGreaterThan(bestRank(straightHand));

    // **4 枚では階段の方が揃いやすいのでフラッシュが上** (four_card_hand_eval.go:9)。
    // フラッシュは最小ベット止まり、同じ手が「ストレート扱い」でも同じ側だが、
    // 順序を逆に写すと下の three-of-a-kind との境界がずれる。
    const flush = base({
      playerHand: [card('SPADE', 2), card('SPADE', 5), card('SPADE', 9), card('SPADE', 12), card('HEART', 3)],
    });
    expect(getFourCardPokerHint(flush)?.targetAction).toBe('play-1');

    const straight = base({
      playerHand: [card('SPADE', 5), card('HEART', 6), card('CLOVER', 7), card('DIAMOND', 8), card('SPADE', 12)],
    });
    expect(getFourCardPokerHint(straight)?.targetAction).toBe('play-1');
  });

  it('pushes the maximum on a straight flush', () => {
    const s = base({
      playerHand: [card('SPADE', 5), card('SPADE', 6), card('SPADE', 7), card('SPADE', 8), card('HEART', 12)],
    });
    expect(getFourCardPokerHint(s)?.targetAction).toBe('play-3');
  });

  it('pushes the maximum on four of a kind', () => {
    const s = base({
      playerHand: [card('SPADE', 9), card('HEART', 9), card('CLOVER', 9), card('DIAMOND', 9), card('SPADE', 2)],
    });
    expect(getFourCardPokerHint(s)?.targetAction).toBe('play-3');
  });

  it('picks the best four of the five cards, not the first four', () => {
    // 先頭 4 枚はバラバラだが、5 枚目を入れるとスリーカードになる。
    const s = base({
      playerHand: [card('SPADE', 7), card('HEART', 7), card('CLOVER', 2), card('DIAMOND', 5), card('SPADE', 7)],
    });
    expect(getFourCardPokerHint(s)?.targetAction).toBe('play-3');
  });

  it('reads a high ace straight', () => {
    const s = base({
      playerHand: [card('SPADE', 11), card('HEART', 12), card('CLOVER', 13), card('DIAMOND', 1), card('SPADE', 3)],
    });
    expect(getFourCardPokerHint(s)?.targetAction).toBe('play-1');
  });

  it('stays quiet before any cards are dealt', () => {
    expect(getFourCardPokerHint(base({ playerHand: [] }))).toBeNull();
  });
});
