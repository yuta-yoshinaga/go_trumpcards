import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, DesmochePlayer, DesmocheResponse } from '../../../types/card';
import { formatDesmocheState } from './desmocheFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<DesmochePlayer>): DesmochePlayer {
  return {
    id,
    isHuman,
    cardCount: 9,
    cards: isHuman ? [card('SPADE', 7), card('HEART', 7)] : [],
    score: -10,
    meldedCount: 3,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<DesmocheResponse>): DesmocheResponse {
  return {
    players: [seat(0, true), seat(1, false)],
    phase: 1,
    currentPlayerIdx: 0,
    stockCount: 15,
    discardTop: card('HEART', 9),
    melds: [
      {
        owner: 1,
        kind: 0,
        cards: [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7)],
      },
    ],
    roundNo: 0,
    pot: 40,
    goOutSize: 10,
    roundWinner: -1,
    roundExhausted: false,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('formatDesmocheState', () => {
  it('prints both rules every frame', () => {
    // 「9枚ではなく10枚」と「ポーカーの役は使わない」が最も誤解されやすい。
    const out = formatDesmocheState(makeState());
    expect(out).toContain('exactly 10 cards');
    expect(out).toContain('poker hand rankings play no part');
  });

  it('names the meld kind rather than printing its number', () => {
    expect(formatDesmocheState(makeState())).toContain('[0] set (seat 1)');
    const run = formatDesmocheState(makeState({ melds: [{ owner: 0, kind: 1, cards: [card('SPADE', 5)] }] }));
    expect(run).toContain('run');
  });

  it('prints the pot and how far each seat is down', () => {
    const out = formatDesmocheState(makeState());
    expect(out).toContain('pot 40');
    expect(out).toContain('3/10 down');
    expect(out).toContain('9 cards'); // the hidden hand is a count
  });

  it('shows an empty discard rather than nothing', () => {
    expect(formatDesmocheState(makeState({ discardTop: undefined }))).toContain('discard: -');
  });

  it('says when nobody got down, so the growing pot is explained', () => {
    const out = formatDesmocheState(makeState({ roundWinner: -1, roundExhausted: true, pot: 80 }));
    expect(out).toContain('nobody got down to ten');
    expect(out).toContain('80');
  });

  it('reports a winner instead when there is one', () => {
    const out = formatDesmocheState(makeState({ roundWinner: 1, roundExhausted: false }));
    expect(out).toContain('seat 1 melded ten');
    expect(out).not.toContain('nobody got down to ten');
  });

  it('reports each ending', () => {
    expect(formatDesmocheState(makeState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('you finish ahead');
    expect(formatDesmocheState(makeState({ gameEndFlag: true, winnerIdx: 2 }))).toContain('you finish behind');
  });
});
