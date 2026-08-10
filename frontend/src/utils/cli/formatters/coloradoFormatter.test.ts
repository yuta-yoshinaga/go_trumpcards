import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, ColoradoResponse } from '../../../types/card';
import { formatColoradoState } from './coloradoFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeState(overrides: Partial<ColoradoResponse> = {}): ColoradoResponse {
  return {
    tableau: [[card('SPADE', 1)], [], [card('HEART', 9), card('CLOVER', 4)]],
    foundation: [[card('SPADE', 1)], [], [], [], [card('SPADE', 13)], [], [], []],
    foundationAscending: [true, true, true, true, false, false, false, false],
    stockCount: 71,
    waste: [card('HEART', 6)],
    phase: 0,
    moveCount: 13,
    canUndo: true,
    message: '',
    ...overrides,
  };
}

describe('formatColoradoState', () => {
  it('marks which foundations build up and which build down', () => {
    const out = formatColoradoState(makeState());
    const line = out.split('\n').find((l) => l.startsWith('foundation:'));
    expect(line).toBeDefined();
    expect(line).toContain('↑');
    expect(line).toContain('↓');
  });

  it('shows the stock count and the waste top', () => {
    const out = formatColoradoState(makeState());
    expect(out).toContain('stock: 71');
    expect(out).toMatch(/waste: .+/);
  });

  it('marks an empty waste rather than dropping the line', () => {
    expect(formatColoradoState(makeState({ waste: [] }))).toContain('waste: [  ]');
  });

  it('lists every pile with its depth', () => {
    const out = formatColoradoState(makeState());
    expect(out).toContain('T0:');
    expect(out).toContain('T1: [  ] (0)');
    expect(out).toContain('T2:');
    expect(out).toContain('(2)');
  });

  it('shows the move count and any message', () => {
    const out = formatColoradoState(makeState({ message: 'no moves left' }));
    expect(out).toContain('moves: 13');
    expect(out).toContain('no moves left');
  });
});
