import { describe, expect, it } from 'vitest';
import type { StreetsAndAlleysResponse } from '../../../types/card';
import { formatStreetsAndAlleysState } from './streetsandalleysFormatter';

function makeState(overrides?: Partial<StreetsAndAlleysResponse>): StreetsAndAlleysResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatStreetsAndAlleysState', () => {
  it('renders header and empty board', () => {
    const result = formatStreetsAndAlleysState(makeState());
    expect(result).toContain('Streets and Alleys');
    expect(result).toContain('foundation:');
    expect(result).toContain('t0: [empty]');
  });

  it('renders cards in tableau', () => {
    const result = formatStreetsAndAlleysState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 1 }, faceUp: true }], ...Array.from({ length: 7 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders foundation top card', () => {
    const result = formatStreetsAndAlleysState(
      makeState({
        foundation: [[{ design: 'HEART', value: 1 }], [], [], []],
      }),
    );
    expect(result).toContain('foundation:');
  });

  it('shows hint when present', () => {
    const result = formatStreetsAndAlleysState(
      makeState({
        hint: { fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 2 },
        messageCode: 'streetsandalleys.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('tableau2');
  });

  it('shows stalemate message', () => {
    const result = formatStreetsAndAlleysState(makeState({ isStalemate: true }));
    expect(result).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    const result = formatStreetsAndAlleysState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 2 };
    expect(formatStreetsAndAlleysState(makeState({ hint, messageCode: 'streetsandalleys.hintAvailable' }))).toContain(
      'HINT:',
    );
    expect(formatStreetsAndAlleysState(makeState({ hint, messageCode: 'streetsandalleys.playing' }))).not.toContain(
      'HINT:',
    );
  });
});
