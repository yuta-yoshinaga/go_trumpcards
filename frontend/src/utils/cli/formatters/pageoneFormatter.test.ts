import { describe, expect, it } from 'vitest';
import type { PageOneResponse } from '../../../types/card';
import { formatPageoneState } from './pageoneFormatter';

const basePlayer = {
  id: 0,
  isHuman: true,
  cardCount: 2,
  cards: [
    { design: 'SPADE' as const, value: 5 },
    { design: 'CLOVER' as const, value: 10 },
  ],
  roundScore: 0,
  cumulativeScore: 0,
  hasDeclared: false,
};

function makeState(overrides: Partial<PageOneResponse> = {}): PageOneResponse {
  return {
    players: [basePlayer, { ...basePlayer, id: 1, isHuman: false, cards: [], cardCount: 3 }],
    phase: 0,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: { design: 'SPADE', value: 7 },
    drawPileCount: 30,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 200 },
    ...overrides,
  };
}

describe('formatPageoneState', () => {
  it('renders basic play phase', () => {
    const text = formatPageoneState(makeState());
    expect(text).toContain('Page One');
    expect(text).toContain('round: 1');
    expect(text).toContain('draw pile: 30');
    expect(text).toContain('discard:');
  });

  it('shows declared badge when a player has declared', () => {
    const text = formatPageoneState(
      makeState({
        players: [{ ...basePlayer, hasDeclared: true, cardCount: 1 }],
      }),
    );
    expect(text).toContain('[PAGE ONE!]');
  });

  it('shows must-declare prompt in phase 1', () => {
    const text = formatPageoneState(makeState({ phase: 1 }));
    expect(text).toContain('Declare Page One');
  });

  it('shows winner when game ends', () => {
    const text = formatPageoneState(makeState({ gameEndFlag: true, winnerIdx: 0 }));
    expect(text).toContain('Game Over');
    expect(text).toContain('Winner');
  });

  it('includes custom message when present', () => {
    const text = formatPageoneState(makeState({ message: 'invalid move' }));
    expect(text).toContain('invalid move');
  });

  it('omits discard line when discardTop is null', () => {
    const text = formatPageoneState(makeState({ discardTop: null }));
    expect(text).not.toContain('discard:');
  });
});
