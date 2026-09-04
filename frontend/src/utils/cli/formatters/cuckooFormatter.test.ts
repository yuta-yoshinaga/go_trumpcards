import { describe, expect, it } from 'vitest';
import type { CuckooPlayer, CuckooResponse } from '../../../types/card';
import { formatCuckooState } from './cuckooFormatter';

function makePlayer(overrides: Partial<CuckooPlayer> = {}): CuckooPlayer {
  return {
    id: 1,
    isHuman: false,
    card: null,
    lives: 3,
    isEliminated: false,
    kingRevealed: false,
    isCurrentTurn: false,
    ...overrides,
  };
}

function makeState(overrides: Partial<CuckooResponse> = {}): CuckooResponse {
  return {
    players: [
      makePlayer({ id: 0, isHuman: true, card: { design: 'SPADE', value: 5 }, isCurrentTurn: true }),
      makePlayer({ id: 1 }),
      makePlayer({ id: 2 }),
      makePlayer({ id: 3 }),
    ],
    phase: 0,
    roundNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    stockCount: 47,
    gameEndFlag: false,
    winnerIdx: -1,
    pendingSwapFrom: -1,
    pendingSwapTo: -1,
    roundLowest: -1,
    roundLosers: [],
    swapTargetIdx: 1,
    config: { cpuDifficulty: 1, initialLives: 3 },
    message: '',
    ...overrides,
  };
}

describe('formatCuckooState', () => {
  it('includes the header, round, phase, dealer and stock', () => {
    const out = formatCuckooState(makeState());
    expect(out).toContain('Cuckoo');
    expect(out).toContain('round: 1');
    expect(out).toContain('phase: Turn');
    expect(out).toContain('stock: 47');
  });

  it('renders lives as hearts and shows the turn prompt', () => {
    const out = formatCuckooState(makeState());
    expect(out).toContain('♥♥♥');
    expect(out).toContain('your turn');
  });

  it('shows the refuse prompt when the human is the swap target', () => {
    const out = formatCuckooState(makeState({ phase: 1, pendingSwapFrom: 3, pendingSwapTo: 0 }));
    expect(out).toContain('refuse');
  });

  it('shows eliminated and round-loser info', () => {
    const out = formatCuckooState(
      makeState({
        phase: 2,
        roundLowest: 2,
        roundLosers: [1],
        players: [
          makePlayer({ id: 0, isHuman: true, card: { design: 'SPADE', value: 5 } }),
          makePlayer({ id: 1, lives: 0, isEliminated: true }),
          makePlayer({ id: 2 }),
          makePlayer({ id: 3 }),
        ],
      }),
    );
    expect(out).toContain('(out)');
    expect(out).toContain('lost a life');
    expect(out).toContain('round over');
  });

  it('announces the winner on game end', () => {
    const out = formatCuckooState(makeState({ phase: 3, gameEndFlag: true, winnerIdx: 0 }));
    expect(out).toContain('Game Over');
  });
});
