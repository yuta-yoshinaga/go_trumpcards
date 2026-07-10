import { describe, expect, it } from 'vitest';
import type { IndianRummyResponse } from '../../../types/card';
import { formatIndianrummyState } from './indianrummyFormatter';

function makeState(overrides: Partial<IndianRummyResponse> = {}): IndianRummyResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 13,
        cards: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 10 },
        ],
        roundScore: 0,
        cumulativeScore: 12,
        deadwood: 20,
        hasPureSequence: true,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 13,
        cards: [],
        roundScore: 5,
        cumulativeScore: 40,
        deadwood: 35,
        hasPureSequence: false,
      },
    ],
    phase: 0,
    roundNumber: 2,
    targetRounds: 5,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    discardTop: { design: 'CLOVER', value: 7 },
    drawPileCount: 40,
    wildJoker: { design: 'DIAMOND', value: 5 },
    wildRank: 5,
    gameEndFlag: false,
    winnerIdx: -1,
    declarerIdx: -1,
    declarationValid: false,
    config: { playerCount: 2, cpuDifficulty: 1, targetRounds: 5 },
    message: '',
    messageCode: '',
    messageParams: {},
    ...overrides,
  } as IndianRummyResponse;
}

describe('formatIndianrummyState', () => {
  it('renders the header, round, phase, wild joker and discard', () => {
    const out = formatIndianrummyState(makeState());
    expect(out).toContain('Indian Rummy');
    expect(out).toContain('round: 2/5');
    expect(out).toContain('phase: DRAW');
    expect(out).toContain('wild joker:');
    expect(out).toContain('stock: 40');
  });

  it('renders each player and the human hand', () => {
    const out = formatIndianrummyState(makeState());
    expect(out).toContain('total=12');
    expect(out).toContain('round=5');
    expect(out).toContain('cards=13');
  });

  it('shows placeholders when wild joker and discard are absent', () => {
    const out = formatIndianrummyState(makeState({ wildJoker: null, discardTop: null }));
    expect(out).toContain('wild joker: [  ]');
    expect(out).toContain('discard: [  ]');
  });

  it('falls back to UNKNOWN for an out-of-range phase', () => {
    const out = formatIndianrummyState(makeState({ phase: 9 }));
    expect(out).toContain('phase: UNKNOWN');
  });

  it('appends the message and the winner line at game end', () => {
    const out = formatIndianrummyState(makeState({ gameEndFlag: true, winnerIdx: 0, message: 'You win!' }));
    expect(out).toContain('You win!');
    expect(out).toContain('Game Over! Winner:');
  });
});
