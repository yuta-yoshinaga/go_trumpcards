import { describe, expect, it } from 'vitest';
import type { Card, GongZhuResponse } from '../../../types/card';
import { formatGongZhuState } from './gongzhuFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<GongZhuResponse> = {}): GongZhuResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [card('SPADE', 12), card('DIAMOND', 11)],
        capturedPointCards: [],
        roundScore: -10,
        cumulativeScore: -30,
        trickCount: 1,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 2,
        cards: [],
        capturedPointCards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 2,
        isHuman: false,
        cardCount: 2,
        cards: [],
        capturedPointCards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 3,
        isHuman: false,
        cardCount: 2,
        cards: [],
        capturedPointCards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
    ],
    phase: 1,
    roundNumber: 1,
    trickNumber: 2,
    currentPlayerIdx: 0,
    currentTrick: [],
    heartsBroken: false,
    exposed: { pig: false, sheep: false, ace: false, doubler: false },
    exposableIndices: [],
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 0, pointLimit: 1000 },
    ...overrides,
  };
}

describe('formatGongZhuState', () => {
  it('renders header, round, and exposure summary (none)', () => {
    const out = formatGongZhuState(makeState());
    expect(out).toContain('Gong Zhu');
    expect(out).toContain('exposed: none');
  });

  it('renders exposed point cards', () => {
    const out = formatGongZhuState(makeState({ exposed: { pig: true, sheep: true, ace: false, doubler: false } }));
    expect(out).toContain('♠Q');
    expect(out).toContain('♦J');
  });

  it('shows human cards and player scores', () => {
    const out = formatGongZhuState(makeState());
    expect(out).toContain('total=-30');
    expect(out).toContain('tricks=1');
  });

  it('renders the current trick when present', () => {
    const out = formatGongZhuState(makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 7) }] }));
    expect(out).toContain('trick:');
  });

  it('shows exposure prompt during expose phase', () => {
    const out = formatGongZhuState(makeState({ phase: 0 }));
    expect(out).toContain('Exposure phase');
  });

  it('renders hint and game over', () => {
    const out = formatGongZhuState(
      makeState({
        hint: { cardIndices: [0], reason: 'discard_pig' },
        messageCode: 'gongzhu.hintRequested',
        gameEndFlag: true,
        winnerIdx: 1,
        message: 'done',
      }),
    );
    expect(out).toContain('HINT');
    expect(out).toContain('Game Over');
    expect(out).toContain('done');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [0], reason: 'discard_pig' };
    expect(formatGongZhuState(makeState({ hint, messageCode: 'gongzhu.hintRequested' }))).toContain('HINT');
    expect(formatGongZhuState(makeState({ hint, messageCode: 'gongzhu.playing' }))).not.toContain('HINT');
  });
});
