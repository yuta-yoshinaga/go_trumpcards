import { describe, expect, it } from 'vitest';
import type { Card, TuSacResponse } from '../../../types/card';
import { TuSacPhase } from '../../../types/phases';
import { formatTuSacState } from './tusacFormatter';

const card = (glyph: string, color: string, value: number): Card =>
  ({ design: 'SPADE', value, glyph, label: glyph, color, deck: 'tusac' }) as Card;

const seat = (over: Partial<TuSacResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    cards: [card('卒', 'red', 7), card('車', 'green', 4)],
    handCount: 2,
    melds: [],
    meldPoints: 0,
    score: 0,
    roundScore: 0,
    isTurn: true,
    wentOut: false,
    ...over,
  }) as TuSacResponse['seats'][number];

const base = {
  phase: TuSacPhase.DRAW,
  seats: [seat(), seat({ name: 'CPU1', isHuman: false, cards: [], handCount: 20, isTurn: false })],
  discardTop: card('象', 'gold', 3),
  discardCount: 1,
  stockCount: 31,
  turnSeat: 0,
  humanSeat: 0,
  isHumanTurn: true,
  roundNumber: 2,
  rounds: 5,
  wentOutSeat: -1,
  handSize: 20,
  deckSize: 112,
  meldPointsByKind: [0, 2, 3, 5],
  winnerSeat: 0,
  gameEndFlag: false,
  message: '',
} as TuSacResponse;

const withState = (over: Partial<TuSacResponse>): TuSacResponse => ({ ...base, ...over });

describe('formatTuSacState', () => {
  it('フェーズ・ラウンド・山を出す', () => {
    const out = formatTuSacState(base);
    expect(out).toContain('Phase: DRAW');
    expect(out).toContain('Round: 2 of 5');
    expect(out).toContain('Stock: 31');
  });

  // **札は色 + 駒で書く。** ランクとスートではない。
  it('色と駒で札を書く', () => {
    const out = formatTuSacState(base);
    // 赤の卒、緑の車、黄(gold)の象。
    expect(out).toContain('R卒');
    expect(out).toContain('G車');
    expect(out).toContain('Y象');
  });

  // **番号は 1 始まり。** 同じ札が 4 枚あるので名前では指定できない。
  it('手札に 1 始まりの番号を振る', () => {
    const out = formatTuSacState(base);
    expect(out).toContain('1:R卒');
    expect(out).toContain('2:G車');
    expect(out).not.toContain('0:R卒');
  });

  // **相手の手札は届かない。** 枚数と場だけが出る。
  it('相手は枚数だけを出す', () => {
    const out = formatTuSacState(base);
    expect(out).toContain('CPU1 20 cards');
    // 自分の手札の行は 1 つだけ。
    expect(out.split('Your hand:').length - 1).toBe(1);
  });

  it('場の組み合わせを出す', () => {
    const out = formatTuSacState(
      withState({
        seats: [
          seat(),
          seat({
            name: 'CPU1',
            isHuman: false,
            cards: [],
            handCount: 15,
            isTurn: false,
            melds: [{ kind: 3, points: 5, cards: [card('卒', 'red', 7)] }],
            meldPoints: 5,
          }),
        ],
      }),
    );
    expect(out).toContain('five soldiers');
  });

  it('場面ごとに促す操作を変える', () => {
    expect(formatTuSacState(base)).toContain('Draw (draw) or take the discard');
    expect(formatTuSacState(withState({ phase: TuSacPhase.DISCARD }))).toContain('Meld or discard');
  });

  it('決着で得点と勝者を出す', () => {
    const out = formatTuSacState(
      withState({
        phase: TuSacPhase.GAME_END,
        gameEndFlag: true,
        winnerSeat: 1,
        seats: [
          seat({ handCount: 3, meldPoints: 5, roundScore: 2 }),
          seat({
            name: 'CPU1',
            isHuman: false,
            cards: [],
            handCount: 0,
            meldPoints: 12,
            roundScore: 12,
            wentOut: true,
          }),
        ],
      }),
    );
    expect(out).toContain('melds 5 - held 3 = 2');
    expect(out).toContain('(went out)');
    expect(out).toContain('Winner: CPU1');
  });

  it('捨て札が空でも落ちない', () => {
    const out = formatTuSacState(withState({ discardTop: null, discardCount: 0 }));
    expect(out).toContain('Top discard: -');
  });
});
