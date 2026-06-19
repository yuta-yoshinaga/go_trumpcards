import { describe, expect, it } from 'vitest';
import type { Card, PishtiPlayer, PishtiResponse } from '../../../types/card';
import { formatPishtiState } from './pishtiFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makePlayer(overrides: Partial<PishtiPlayer> = {}): PishtiPlayer {
  return {
    id: 1,
    isHuman: false,
    cardCount: 4,
    cards: [],
    capturedCount: 0,
    pistiBonus: 0,
    finalScore: 0,
    ...overrides,
  };
}

function makeState(overrides: Partial<PishtiResponse> = {}): PishtiResponse {
  return {
    players: [
      makePlayer({ id: 0, isHuman: true, cards: [card('SPADE', 5), card('HEART', 11)] }),
      makePlayer({ id: 1 }),
      makePlayer({ id: 2 }),
      makePlayer({ id: 3 }),
    ],
    currentTurn: 0,
    pile: [card('CLOVER', 7)],
    pileTop: card('CLOVER', 7),
    pileCount: 1,
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 'play',
    remainingDeck: 36,
    winners: [],
    finalScores: [],
    config: { playerCnt: 4, cpuDifficulty: 1 },
    message: '',
    ...overrides,
  };
}

describe('formatPishtiState', () => {
  it('includes the header, phase, stock and pile top', () => {
    const out = formatPishtiState(makeState());
    expect(out).toContain('Pişti');
    expect(out).toContain('phase: Play');
    expect(out).toContain('stock: 36');
    expect(out).toContain('pile: top=');
  });

  it('renders captured/Pişti lines and the turn prompt', () => {
    const out = formatPishtiState(makeState());
    expect(out).toContain('captured 0');
    expect(out).toContain('Pişti 0');
    expect(out).toContain('your turn');
    expect(out).toContain('your hand');
  });

  it('shows an empty pile top as a dash', () => {
    const out = formatPishtiState(makeState({ pile: [], pileTop: null, pileCount: 0 }));
    expect(out).toContain('pile: top=- (0 cards)');
  });

  it('announces the winner and final scores on game end', () => {
    const out = formatPishtiState(
      makeState({
        phase: 'gameEnd',
        gameEndFlag: true,
        currentTurn: -1,
        winners: [0],
        finalScores: [11, 7, 4, 8],
      }),
    );
    expect(out).toContain('Game Over');
    expect(out).toContain('11 pts');
    expect(out).toContain('Winner');
  });
});
