import { describe, expect, it } from 'vitest';
import type { CruelResponse } from '../../../types/card';
import { formatCruelState } from './cruelFormatter';

const card = (design: string, value: number) => ({ design, value }) as CruelResponse['foundation'][number][number];

function makeState(overrides: Partial<CruelResponse> = {}): CruelResponse {
  return {
    foundation: [[card('SPADE', 1)], [], [], []],
    tableau: [
      [
        { faceUp: false, card: null },
        { faceUp: true, card: card('HEART', 13) },
      ],
      [],
    ],
    moveCount: 4,
    phase: 1,
    ...overrides,
  } as CruelResponse;
}

describe('formatCruelState', () => {
  it('renders foundations, tableau, and footer', () => {
    const out = formatCruelState(makeState());
    expect(out).toContain('♠: SPADE-1 (1)');
    expect(out).toContain('♣: empty (0)');
    expect(out).toContain('Moves: 4  Phase: 1');
  });

  it('hides face-down cards and marks empty columns', () => {
    const out = formatCruelState(makeState());
    expect(out).toContain('[0]??');
    expect(out).toContain('[1]HEART-13');
    expect(out).toContain('(empty)');
  });
});
