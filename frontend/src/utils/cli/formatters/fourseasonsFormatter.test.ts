import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, FourSeasonsResponse } from '../../../types/card';
import { formatFourSeasonsState } from './fourseasonsFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeState(overrides: Partial<FourSeasonsResponse> = {}): FourSeasonsResponse {
  return {
    tableau: [[card('HEART', 12)], [], [card('DIAMOND', 9), card('CLOVER', 8)], [], []],
    foundation: [[card('SPADE', 7)], [], [], []],
    stockCount: 44,
    waste: [card('HEART', 6)],
    baseRank: 7,
    phase: 0,
    moveCount: 3,
    canUndo: true,
    message: '',
    ...overrides,
  };
}

describe('formatFourSeasonsState', () => {
  it('prints the base rank next to the foundations', () => {
    // Reading the corners without the base rank tells you nothing, because
    // every corner builds from it.
    expect(formatFourSeasonsState(makeState())).toContain('base: 7');
  });

  it('prints five cross piles with their depth', () => {
    const out = formatFourSeasonsState(makeState());
    expect(out).toContain('T0:');
    expect(out).toContain('T4:');
    expect(out).toContain('(2)');
  });

  it('shows an empty marker for empty piles rather than dropping them', () => {
    const out = formatFourSeasonsState(makeState({ waste: [], foundation: [[], [], [], []] }));
    expect(out).toContain('[  ]');
  });

  it('prints the move count', () => {
    expect(formatFourSeasonsState(makeState())).toContain('moves: 3');
  });

  it('appends the server message when there is one', () => {
    expect(formatFourSeasonsState(makeState({ message: 'ゲームクリア！' }))).toContain('ゲームクリア！');
  });

  it('omits the message line when empty', () => {
    expect(formatFourSeasonsState(makeState({ message: '' })).endsWith('moves: 3')).toBe(true);
  });
});
