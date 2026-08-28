import { describe, expect, it } from 'vitest';
import { actionDesc, isCardPlayable, isPositionPlayable, listJokerPlacements, wrapValue } from './sevensUtils';

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

describe('listJokerPlacements', () => {
  it('returns the 6 and 8 positions on each suit when only 7 is placed', () => {
    // Standard opening: every suit has only its 7 placed.
    const board7Only = [0, 128, 128, 128, 128];
    const slots = listJokerPlacements(board7Only, false, false, 0);
    // 4 suits * 2 neighbours (6 and 8) = 8 placements
    expect(slots).toHaveLength(8);
    // Spot-check that suit 1 (♠) offers values 6 and 8.
    const spadeValues = slots
      .filter((s) => s.suit === 1)
      .map((s) => s.value)
      .sort((a, b) => a - b);
    expect(spadeValues).toEqual([6, 8]);
  });

  it('returns a single slot when only one suit has been opened', () => {
    // Bit 7 set on spade only.
    const board = [0, 128, 0, 0, 0];
    const slots = listJokerPlacements(board, false, false, 0);
    expect(slots).toHaveLength(2); // 6 and 8 of spades
  });

  it('returns an empty list when the board is empty (no anchors to extend)', () => {
    const empty = [0, 0, 0, 0, 0];
    expect(listJokerPlacements(empty, false, false, 0)).toEqual([]);
  });
});

describe('actionDesc joker reclaim', () => {
  const players = [
    { id: 0, isHuman: true },
    { id: 1, isHuman: false },
  ];
  // Echo the key and its params so the assertion reads what was actually asked
  // for, rather than a translated string that could resolve to anything.
  const t = (key: string, opts?: Record<string, unknown>) => `${key}${opts ? `(${Object.keys(opts).join(',')})` : ''}`;

  it('mentions the reclaim, which is otherwise a silent extra card', () => {
    const desc = actionDesc(
      players,
      {
        playerIdx: 0,
        playedCard: { design: 'SPADE', value: 6 },
        targetSuit: 0,
        targetValue: 0,
        forcedPass: false,
        jokerReclaimed: true,
      },
      t,
    );
    expect(desc).toContain('actionPlayed');
    expect(desc).toContain('actionReclaimedJoker');
  });

  it('says nothing about a reclaim that did not happen', () => {
    const desc = actionDesc(
      players,
      {
        playerIdx: 0,
        playedCard: { design: 'SPADE', value: 6 },
        targetSuit: 0,
        targetValue: 0,
        forcedPass: false,
        jokerReclaimed: false,
      },
      t,
    );
    expect(desc).not.toContain('actionReclaimedJoker');
  });
});
