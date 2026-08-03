import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, LobaPlayer, LobaResponse } from '../../../types/card';
import { formatLobaState } from './lobaFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<LobaPlayer>): LobaPlayer {
  return {
    id,
    isHuman,
    cardCount: 9,
    cards: isHuman ? [card('SPADE', 7), card('HEART', 7)] : [],
    score: 12,
    eliminated: false,
    hasMelded: false,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<LobaResponse>): LobaResponse {
  return {
    players: [seat(0, true), seat(1, false)],
    phase: 1,
    currentPlayerIdx: 0,
    stockCount: 70,
    discardTop: card('HEART', 9),
    melds: [
      {
        owner: 1,
        kind: 0,
        cards: [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7)],
      },
    ],
    roundNo: 0,
    knockOut: 101,
    roundWinner: -1,
    roundClean: false,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('formatLobaState', () => {
  it('prints both rules every frame', () => {
    // 「異なる3スート」と「ジョーカーは並びだけ」が最も間違えやすい。
    const out = formatLobaState(makeState());
    expect(out).toContain('three DIFFERENT suits');
    expect(out).toContain('never in a pierna');
  });

  it('names the meld kind rather than printing its number', () => {
    expect(formatLobaState(makeState())).toContain('[0] pierna (seat 1)');
    const run = formatLobaState(makeState({ melds: [{ owner: 0, kind: 1, cards: [card('SPADE', 5)] }] }));
    expect(run).toContain('escalera');
  });

  it('prints the knock-out threshold and every score', () => {
    const out = formatLobaState(makeState());
    expect(out).toContain('out at 101');
    expect(out).toContain('12 penalty');
    expect(out).toContain('9 cards'); // the hidden hand is a count
  });

  it('marks an eliminated seat', () => {
    const out = formatLobaState(makeState({ players: [seat(0, true), seat(1, false, { eliminated: true })] }));
    expect(out).toContain('-- out');
  });

  it('shows an empty discard rather than nothing', () => {
    expect(formatLobaState(makeState({ discardTop: undefined }))).toContain('discard: -');
  });

  it('tells a clean go-out apart', () => {
    // -10 が付くかどうかは表示から読めなければならない。
    expect(formatLobaState(makeState({ roundWinner: 1 }))).toContain('seat 1 went out');
    expect(formatLobaState(makeState({ roundWinner: 1, roundClean: true }))).toContain('in one go (-10)');
  });

  it('reports each ending', () => {
    expect(formatLobaState(makeState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('last one standing');
    expect(formatLobaState(makeState({ gameEndFlag: true, winnerIdx: 2 }))).toContain('knocked out');
  });
});
