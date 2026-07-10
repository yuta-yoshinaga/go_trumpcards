import { describe, expect, it } from 'vitest';
import type { PanResponse } from '../../../types/card';
import { formatPanState } from './panFormatter';

function makeState(overrides: Partial<PanResponse> = {}): PanResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 11,
        cards: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 7 },
        ],
        laidMelds: [{ cards: [{ design: 'SPADE', value: 3 }] }],
        meldedCount: 3,
        chips: 50,
        handPoints: 12,
        roundScore: 0,
        cumulativeScore: 12,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 10,
        cards: [],
        laidMelds: [],
        meldedCount: 0,
        chips: 45,
        handPoints: 20,
        roundScore: 5,
        cumulativeScore: 40,
      },
    ],
    phase: 1,
    roundNumber: 2,
    targetRounds: 5,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    discardTop: { design: 'CLOVER', value: 7 },
    drawPileCount: 250,
    deckSize: 320,
    winMeldCount: 11,
    gameEndFlag: false,
    winnerIdx: -1,
    panDeclarerIdx: -1,
    config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 5 },
    message: '',
    messageCode: '',
    messageParams: {},
    ...overrides,
  } as PanResponse;
}

describe('formatPanState', () => {
  it('renders the header, round, phase, and discard', () => {
    const out = formatPanState(makeState());
    expect(out).toContain('Panguingue / Pan');
    expect(out).toContain('round: 2/5');
    expect(out).toContain('phase: PLAY');
    expect(out).toContain('stock: 250');
  });

  it('renders each player with chips, melded count, and laid melds', () => {
    const out = formatPanState(makeState());
    expect(out).toContain('chips=50');
    expect(out).toContain('melded=3');
    expect(out).toContain('meld[0]:');
  });

  it('shows a placeholder when the discard is absent', () => {
    const out = formatPanState(makeState({ discardTop: null }));
    expect(out).toContain('discard: [  ]');
  });

  it('falls back to UNKNOWN for an out-of-range phase', () => {
    const out = formatPanState(makeState({ phase: 9 }));
    expect(out).toContain('phase: UNKNOWN');
  });

  it('appends the message and the winner line at game end', () => {
    const out = formatPanState(makeState({ gameEndFlag: true, winnerIdx: 0, message: 'You win!' }));
    expect(out).toContain('You win!');
    expect(out).toContain('Game Over! Winner:');
  });
});
