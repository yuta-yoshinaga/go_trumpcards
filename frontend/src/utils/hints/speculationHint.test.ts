import { describe, expect, it } from 'vitest';
import type { SpeculationResponse } from '../../types/card';
import { SpeculationPhase } from '../../types/phases';
import { getSpeculationHint } from './speculationHint';

const seat = (hiddenCount: number, name = 'CPU') => ({ name, chips: 200, hiddenCount });

const base: SpeculationResponse = {
  phase: SpeculationPhase.FLIP,
  seats: [seat(3, 'You'), seat(3, 'CPU1'), seat(3, 'CPU2'), seat(3, 'CPU3')],
  trumpSuit: 3,
  pot: 40,
  turnSeat: 0,
  bestSeat: -1,
  offerFrom: -1,
  offerTo: -1,
  offerAmount: 0,
  roundNo: 0,
  winnerSeat: -1,
  gameEndFlag: false,
  message: '',
};

const withState = (over: Partial<SpeculationResponse>): SpeculationResponse => ({ ...base, ...over });

/** Auction with the human as the card's owner: the offer is addressed to seat 0. */
const sellingTo = (hiddenPerSeat: number) =>
  withState({
    phase: SpeculationPhase.AUCTION,
    seats: [seat(hiddenPerSeat, 'You'), seat(hiddenPerSeat, 'CPU1'), seat(hiddenPerSeat, 'CPU2')],
    offerFrom: 2,
    offerTo: 0,
    offerAmount: 25,
    bestSeat: 0,
  });

/** Auction with a CPU as the card's owner: the human is the would-be buyer. */
const buyingFrom = (hiddenPerSeat: number) =>
  withState({
    phase: SpeculationPhase.AUCTION,
    seats: [seat(hiddenPerSeat, 'You'), seat(hiddenPerSeat, 'CPU1'), seat(hiddenPerSeat, 'CPU2')],
    offerFrom: 0,
    offerTo: 2,
    offerAmount: 25,
    bestSeat: 2,
  });

describe('getSpeculationHint', () => {
  it('めくりフェーズではめくるよう勧める', () => {
    expect(getSpeculationHint(base)).toEqual({
      targetAction: 'flip',
      reason: 'frontendHint.speculationFlip',
      confidence: 'moderate',
    });
  });

  it('ゲーム終了後は助言しない', () => {
    expect(getSpeculationHint(withState({ gameEndFlag: true }))).toBeNull();
  });

  it('決着フェーズでは助言しない', () => {
    expect(getSpeculationHint(withState({ phase: SpeculationPhase.RESULT }))).toBeNull();
  });

  // **伏せ札が多く残っているうちは、どんな強い札も上を出される。**
  // CUI の speculationAuctionHintKey と同じ判断 (残り > 席数 なら売る)。
  it('売り手で伏せ札が多く残っていれば売るよう勧める', () => {
    // 3 席 x 3 枚 = 9 枚 > 3 席
    expect(getSpeculationHint(sellingTo(3))).toEqual({
      targetAction: 'accept',
      reason: 'frontendHint.speculationSell',
      confidence: 'moderate',
    });
  });

  it('売り手で残りが少なければ断って持ち続けるよう勧める', () => {
    // 3 席 x 1 枚 = 3 枚 <= 3 席
    expect(getSpeculationHint(sellingTo(1))).toEqual({
      targetAction: 'decline',
      reason: 'frontendHint.speculationHold',
      confidence: 'strong',
    });
  });

  it('買い手で残りが少なければ買うよう勧める', () => {
    expect(getSpeculationHint(buyingFrom(1))).toEqual({
      targetAction: 'accept',
      reason: 'frontendHint.speculationBuy',
      confidence: 'moderate',
    });
  });

  it('買い手で伏せ札が多く残っていれば見送るよう勧める', () => {
    expect(getSpeculationHint(buyingFrom(3))).toEqual({
      targetAction: 'decline',
      reason: 'frontendHint.speculationPass',
      confidence: 'strong',
    });
  });

  // **売り手と買い手を取り違えると助言が丸ごと反転する。** 同じ残り枚数で
  // 反対の結論が出ることを、両側から踏んで固定する。
  it('同じ残り枚数でも売り手と買い手で結論が反対になる', () => {
    expect(getSpeculationHint(sellingTo(1))?.reason).toBe('frontendHint.speculationHold');
    expect(getSpeculationHint(buyingFrom(1))?.reason).toBe('frontendHint.speculationBuy');
    expect(getSpeculationHint(sellingTo(3))?.reason).toBe('frontendHint.speculationSell');
    expect(getSpeculationHint(buyingFrom(3))?.reason).toBe('frontendHint.speculationPass');
  });

  it('席が空なら競り中でも助言しない', () => {
    expect(getSpeculationHint(withState({ phase: SpeculationPhase.AUCTION, seats: [] }))).toBeNull();
  });
});
