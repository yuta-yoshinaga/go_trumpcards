import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, TrexPlayer, TrexResponse } from '../../../types/card';
import { formatTrexState } from './trexFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<TrexPlayer>): TrexPlayer {
  return {
    id,
    isHuman,
    cardCount: 13,
    cards: isHuman ? [card('SPADE', 11), card('HEART', 1)] : [],
    score: -40,
    dealScore: -10,
    tricksWon: 1,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<TrexResponse>): TrexResponse {
  return {
    players: [seat(0, true), seat(1, false)],
    phase: 1,
    currentPlayerIdx: 0,
    kingIdx: 0,
    contract: 2,
    availableContracts: [3, 4],
    isTrix: false,
    dealNo: 3,
    totalDeals: 20,
    trick: [{ playerIdx: 0, card: card('SPADE', 12) }],
    trickNo: 2,
    runs: [
      { suit: 1, started: true, low: 11, high: 13 },
      { suit: 2, started: false, low: 0, high: 0 },
      { suit: 3, started: false, low: 0, high: 0 },
      { suit: 4, started: false, low: 0, high: 0 },
    ],
    finishOrder: [],
    validIndices: [0],
    canPass: false,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('formatTrexState', () => {
  it('prints both rules every frame', () => {
    // 「1王国に1度だけ」と「ドミノはJ起点」が、このゲームで最も間違えやすい。
    const out = formatTrexState(makeState());
    expect(out).toContain('each contract once per kingdom');
    expect(out).toContain('build out from the JACK');
  });

  it('names the contract rather than printing its number', () => {
    expect(formatTrexState(makeState())).toContain('Queens (-25 each)');
    expect(formatTrexState(makeState({ contract: 5 }))).toContain('not chosen');
  });

  it('lists what is left to choose', () => {
    // 1王国に1度ずつなので、何が残っているかが見えていないと選べない。
    const out = formatTrexState(makeState());
    expect(out).toContain('3=Tricks (-15 each)');
    expect(out).toContain('4=Dominoes');
    expect(formatTrexState(makeState({ availableContracts: [] }))).not.toContain('left to choose');
  });

  it('shows the runs in the dominoes and the trick otherwise', () => {
    const trix = formatTrexState(makeState({ isTrix: true }));
    expect(trix).toContain('spade: 11-13');
    expect(trix).toContain('club: not started');
    expect(trix).not.toContain('trick:');

    const tricks = formatTrexState(makeState());
    expect(tricks).toContain('trick:');
    expect(tricks).not.toContain('spade: 11-13');
  });

  it('marks the king and prints every score', () => {
    const out = formatTrexState(makeState());
    expect(out).toContain('you (king)');
    expect(out).toContain('-40 total (deal -10)');
    expect(out).toContain('13 cards'); // the hidden hand is a count
  });

  it('tells you when a pass is the only move', () => {
    expect(formatTrexState(makeState({ isTrix: true, canPass: true }))).toContain('use s to pass');
    expect(formatTrexState(makeState())).not.toContain('use s to pass');
  });

  it('reports each ending', () => {
    expect(formatTrexState(makeState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('you win');
    expect(formatTrexState(makeState({ gameEndFlag: true, winnerIdx: 2 }))).toContain('you lose');
  });
});
