import { describe, expect, it } from 'vitest';
import { madrassoTrickPoints } from './madrassoTrickPoints';

describe('madrassoTrickPoints', () => {
  it('returns the matching points for each rank and zero for other cards', () => {
    expect(
      madrassoTrickPoints(
        [1, 3, 13, 12, 11, 2].map((value) => ({
          playerIdx: 0,
          card: { design: 'SPADE' as const, value },
        })),
      ),
    ).toBe(30);
  });

  it('returns zero for an empty trick', () => {
    expect(madrassoTrickPoints([])).toBe(0);
  });
});
