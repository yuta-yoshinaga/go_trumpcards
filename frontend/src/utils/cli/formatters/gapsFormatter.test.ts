import { describe, expect, it } from 'vitest';
import type { GapsResponse } from '../../../types/card';
import { formatGapsState } from './gapsFormatter';

function makeState(overrides: Partial<GapsResponse> = {}): GapsResponse {
  const grid: GapsResponse['grid'] = Array.from({ length: 4 }, () => Array(13).fill(null));
  const [firstRow, secondRow] = grid;
  if (firstRow) firstRow[0] = { design: 'SPADE', value: 2 };
  if (secondRow) secondRow[2] = { design: 'HEART', value: 5 };
  return {
    grid,
    lockedPrefixLengths: [1, 0, 0, 0],
    redealsUsed: 0,
    redealsRemaining: 3,
    phase: 0,
    moveCount: 5,
    canUndo: true,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatGapsState', () => {
  it('shows phase, board cards with row-major indices, locked status and next actions', () => {
    const text = formatGapsState(makeState());
    expect(text).toContain('phase: Playing');
    expect(text).toContain('[0] x ♠2 at 0,0 (locked)');
    expect(text).toContain('[15] ♥5 at 1,2');
    expect(text).toContain('p <idx> <row> <col>');
  });

  it('prints a requested server hint, including the returned message code case', () => {
    const state = makeState({
      hint: { fromRow: 1, fromCol: 2, toRow: 0, toCol: 12 },
      messageCode: 'gaps.hintAvailable',
    });
    expect(formatGapsState(state)).toContain('HINT: move card at 1,2 to 0,12');
  });

  it('does not reveal a passive hint without a requested-hint message code', () => {
    const hint = { fromRow: 1, fromCol: 2, toRow: 0, toCol: 12 };
    expect(formatGapsState(makeState({ hint }))).not.toContain('HINT');
    expect(formatGapsState(makeState({ hint, messageCode: 'gaps.playing' }))).not.toContain('HINT');
  });

  it('formats end phases and stalemate advice', () => {
    expect(formatGapsState(makeState({ phase: 1 })).toLowerCase()).toContain('game clear');
    expect(formatGapsState(makeState({ phase: 2 })).toLowerCase()).toContain('game over');
    expect(formatGapsState(makeState({ isStalemate: true, undoToEscape: 4 }))).toContain('Undo 4 moves to escape');
  });
});
