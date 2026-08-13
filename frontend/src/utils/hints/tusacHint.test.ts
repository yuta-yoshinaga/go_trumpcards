import { describe, expect, it } from 'vitest';
import type { Card, TuSacResponse } from '../../types/card';
import { TuSacPhase } from '../../types/phases';
import { getTusacHint } from './tusacHint';

const card = (glyph: string, value: number): Card =>
  ({ design: 'SPADE', value, glyph, label: glyph, color: 'red', deck: 'tusac' }) as Card;

const seat = (over: Partial<TuSacResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    cards: [card('卒', 7), card('車', 4)],
    handCount: 2,
    melds: [],
    meldPoints: 0,
    score: 0,
    roundScore: 0,
    isTurn: true,
    wentOut: false,
    ...over,
  }) as TuSacResponse['seats'][number];

const state = (over: Partial<TuSacResponse> = {}) =>
  ({
    phase: TuSacPhase.DRAW,
    seats: [seat()],
    discardTop: card('象', 3),
    discardCount: 1,
    stockCount: 31,
    turnSeat: 0,
    humanSeat: 0,
    isHumanTurn: true,
    roundNumber: 1,
    rounds: 5,
    wentOutSeat: -1,
    handSize: 20,
    deckSize: 112,
    meldPointsByKind: [0, 2, 3, 5],
    winnerSeat: 0,
    gameEndFlag: false,
    message: '',
    ...over,
  }) as TuSacResponse;

describe('getTusacHint', () => {
  it('終局では助言しない', () => {
    expect(getTusacHint(state({ gameEndFlag: true }))).toBeNull();
  });

  it('決着後は次のラウンドを薦める', () => {
    expect(getTusacHint(state({ phase: TuSacPhase.ROUND_END }))?.targetAction).toBe('next');
  });

  it('他人の手番では助言しない', () => {
    expect(getTusacHint(state({ isHumanTurn: false }))).toBeNull();
  });

  it('引く場面では引くよう薦める', () => {
    expect(getTusacHint(state())?.targetAction).toBe('draw');
  });

  it('捨てる場面では捨てるよう薦める', () => {
    expect(getTusacHint(state({ phase: TuSacPhase.DISCARD }))?.targetAction).toBe('discard');
  });

  // **引いた直後は手札が 1 枚多い。** 捨てて手番を渡す必要がある。
  it('手札が配り枚数を超えていたら理由が変わる', () => {
    const over = getTusacHint(
      state({
        phase: TuSacPhase.DISCARD,
        handSize: 1,
        seats: [seat({ cards: [card('卒', 7), card('車', 4)] })],
      }),
    );
    const normal = getTusacHint(state({ phase: TuSacPhase.DISCARD, handSize: 20 }));
    expect(over?.targetAction).toBe('discard');
    expect(normal?.targetAction).toBe('discard');
    expect(over?.reason).not.toBe(normal?.reason);
  });

  it('席が見つからなければ助言しない', () => {
    expect(getTusacHint(state({ phase: TuSacPhase.DISCARD, humanSeat: 9 }))).toBeNull();
  });
});
