import { describe, expect, it } from 'bun:test';
import { isCardPlayable, isPositionPlayable, wrapValue } from './sevensUtils';

describe('wrapValue', () => {
  it('returns value unchanged for 1-13', () => {
    expect(wrapValue(1)).toBe(1);
    expect(wrapValue(7)).toBe(7);
    expect(wrapValue(13)).toBe(13);
  });

  it('wraps values above 13', () => {
    expect(wrapValue(14)).toBe(1);
    expect(wrapValue(15)).toBe(2);
    expect(wrapValue(26)).toBe(13);
  });

  it('wraps values below 1', () => {
    expect(wrapValue(0)).toBe(13);
    expect(wrapValue(-1)).toBe(12);
    expect(wrapValue(-2)).toBe(11);
  });
});

describe('isPositionPlayable with tunnelSkipWidth', () => {
  // Board with only 7s placed (bit 7 = 128)
  const board7Only = [0, 128, 128, 128, 128];

  it('skip=3: value 4 is playable (7-3=4)', () => {
    expect(isPositionPlayable(board7Only, 1, 4, false, false, 3)).toBe(true);
  });

  it('skip=3: value 10 is playable (7+3=10)', () => {
    expect(isPositionPlayable(board7Only, 1, 10, false, false, 3)).toBe(true);
  });

  it('skip=3: value 5 is not playable', () => {
    expect(isPositionPlayable(board7Only, 1, 5, false, false, 3)).toBe(false);
  });

  it('skip=3: normal adjacency still works (value 6)', () => {
    expect(isPositionPlayable(board7Only, 1, 6, false, false, 3)).toBe(true);
  });

  it('skip=1 is treated as disabled (< 2)', () => {
    expect(isPositionPlayable(board7Only, 1, 5, false, false, 1)).toBe(false);
  });

  it('skip=0 (default) does not add skip connections', () => {
    expect(isPositionPlayable(board7Only, 1, 4, false, false, 0)).toBe(false);
  });

  it('skip=3 without tunnel: no wrap for out-of-range values', () => {
    // Place 1 on board: bit 1 = 2, plus bit 7 = 128 → 130
    const boardWith1 = [0, 130, 128, 128, 128];
    // 1-3 = -2 → out of range without tunnel, so 11 is NOT playable via skip from 1
    expect(isPositionPlayable(boardWith1, 1, 11, false, false, 3)).toBe(false);
  });

  it('skip=3 with tunnel: wraps around for value 1 (low direction)', () => {
    // Place 1 on board
    const boardWith1 = [0, 130, 128, 128, 128];
    // 1-3 → wrapValue(-2) = 11, so 11 IS playable via tunnel wrap
    expect(isPositionPlayable(boardWith1, 1, 11, true, false, 3)).toBe(true);
  });

  it('skip=3 with tunnel: wraps around for value 13 (high direction)', () => {
    // Place 13 on board: bit 13 = 8192, plus bit 7 = 128 → 8320
    const boardWith13 = [0, 8320, 128, 128, 128];
    // 13+3 → wrapValue(16) = 3, so 3 IS playable via tunnel wrap
    expect(isPositionPlayable(boardWith13, 1, 3, true, false, 3)).toBe(true);
  });
});

describe('isCardPlayable with tunnelSkipWidth', () => {
  const board7Only = [0, 128, 128, 128, 128];

  it('passes tunnelSkipWidth through for normal cards', () => {
    const card = { design: 'SPADE' as const, value: 4 };
    expect(isCardPlayable(card, board7Only, false, false, [card], false, false, false, 3)).toBe(true);
    expect(isCardPlayable(card, board7Only, false, false, [card], false, false, false, 0)).toBe(false);
  });

  it('passes tunnelSkipWidth through for joker playability', () => {
    const joker = { design: 'JOKER' as const, value: 0 };
    // With skip=0, board has 6 and 8 playable per suit = 8 positions
    expect(isCardPlayable(joker, board7Only, false, false, [joker], false, false, false, 0)).toBe(true);
    // With skip=3, additional positions are playable
    expect(isCardPlayable(joker, board7Only, false, false, [joker], false, false, false, 3)).toBe(true);
  });
});
