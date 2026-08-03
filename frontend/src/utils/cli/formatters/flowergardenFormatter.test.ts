import { describe, expect, it } from 'vitest';
import type { FlowerGardenResponse } from '../../../types/card';
import { formatFlowerGardenState } from './flowergardenFormatter';

function makeState(overrides?: Partial<FlowerGardenResponse>): FlowerGardenResponse {
  return {
    tableau: Array.from({ length: 6 }, () => []),
    reserve: Array.from({ length: 16 }, () => null),
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatFlowerGardenState', () => {
  it('renders header and empty board', () => {
    const result = formatFlowerGardenState(makeState());
    expect(result).toContain('Flower Garden');
    expect(result).toContain('foundation:');
    expect(result).toContain('reserve:');
    expect(result).toContain('t0: [empty]');
  });

  it('renders cards in tableau', () => {
    const result = formatFlowerGardenState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 1 }, faceUp: true }], ...Array.from({ length: 5 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders reserve cards', () => {
    const result = formatFlowerGardenState(
      makeState({
        reserve: [{ design: 'HEART', value: 7 }, ...Array.from({ length: 15 }, () => null)],
      }),
    );
    expect(result).toContain('reserve:');
  });

  it('renders foundation top card', () => {
    const result = formatFlowerGardenState(
      makeState({
        foundation: [[{ design: 'HEART', value: 1 }], [], [], []],
      }),
    );
    expect(result).toContain('foundation:');
  });

  it('shows hint from tableau when present', () => {
    const result = formatFlowerGardenState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 2 },
        messageCode: 'flowergarden.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t0[1]');
    expect(result).toContain('tableau2');
  });

  // **頼んでいないヒントは CLI に出さない。**#4483 以降 Output() もヒントを載せる。
  it('does not print a passive hint carried on an ordinary response', () => {
    const result = formatFlowerGardenState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 2 },
        messageCode: 'flowergarden.playing',
      }),
    );
    expect(result).not.toContain('HINT');
  });

  it('shows hint from reserve when present', () => {
    const result = formatFlowerGardenState(
      makeState({
        hint: { fromZone: 'reserve', fromCol: 3, cardIndex: 0, toZone: 'foundation', toCol: 1 },
        messageCode: 'flowergarden.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('r3');
  });

  it('shows stalemate message', () => {
    const result = formatFlowerGardenState(makeState({ isStalemate: true }));
    expect(result).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    const result = formatFlowerGardenState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });
});
