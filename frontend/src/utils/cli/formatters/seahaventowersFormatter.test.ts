import { describe, expect, it } from 'vitest';
import type { SeahavenTowersResponse } from '../../../types/card';
import { formatSeahavenTowersState } from './seahaventowersFormatter';

function makeState(overrides?: Partial<SeahavenTowersResponse>): SeahavenTowersResponse {
  return {
    tableau: Array.from({ length: 10 }, () => []),
    reservedCells: [null, null],
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatSeahavenTowersState', () => {
  it('renders header and empty board', () => {
    const result = formatSeahavenTowersState(makeState());
    expect(result).toContain('Seahaven Towers');
    expect(result).toContain('reserved:');
    expect(result).toContain('foundation:');
    expect(result).toContain('col0: [empty]');
  });

  it('renders cards in a tableau column', () => {
    const result = formatSeahavenTowersState(
      makeState({
        tableau: [
          [
            { design: 'SPADE', value: 13 },
            { design: 'SPADE', value: 12 },
          ],
          ...Array.from({ length: 9 }, () => [] as never[]),
        ],
      }),
    );
    expect(result).toContain('[0]');
    expect(result).toContain('[1]');
  });

  it('renders reserved cells with cards', () => {
    const result = formatSeahavenTowersState(
      makeState({
        reservedCells: [{ design: 'HEART', value: 7 }, null],
      }),
    );
    expect(result).toContain('reserved:');
    expect(result).toContain('[c0]');
    expect(result).toContain('[c1]');
  });

  it('renders the top of each foundation pile', () => {
    const result = formatSeahavenTowersState(
      makeState({
        foundation: [
          [
            { design: 'SPADE', value: 1 },
            { design: 'SPADE', value: 2 },
          ],
          [],
          [],
          [],
        ],
      }),
    );
    expect(result).toContain('foundation:');
  });

  it('shows hint when present', () => {
    const result = formatSeahavenTowersState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 2 },
        messageCode: 'seahaventowers.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('tableau');
  });

  it('omits column number from hint when negative', () => {
    const result = formatSeahavenTowersState(
      makeState({
        hint: { fromZone: 'reserved', fromCol: -1, cardIndex: -1, toZone: 'foundation', toCol: -1 },
        messageCode: 'seahaventowers.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).not.toMatch(/reserved-1/);
    expect(result).not.toMatch(/foundation-1/);
  });

  it('shows stalemate message', () => {
    const result = formatSeahavenTowersState(makeState({ isStalemate: true }));
    expect(result).toContain('Stalemate');
  });

  it('echoes the server message when set', () => {
    const result = formatSeahavenTowersState(makeState({ message: 'invalid move' }));
    expect(result).toContain('invalid move');
  });

  it('shows congrats on game-clear phase', () => {
    const result = formatSeahavenTowersState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });

  it('omits congrats on playing phase', () => {
    const result = formatSeahavenTowersState(makeState({ phase: 0 }));
    expect(result).not.toContain('Congratulations');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 1 };
    expect(formatSeahavenTowersState(makeState({ hint, messageCode: 'seahaventowers.hintAvailable' }))).toContain(
      'HINT:',
    );
    expect(formatSeahavenTowersState(makeState({ hint, messageCode: 'seahaventowers.playing' }))).not.toContain(
      'HINT:',
    );
  });
});
