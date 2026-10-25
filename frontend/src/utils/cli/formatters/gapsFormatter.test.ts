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

  it('prints every gap by row and column with the page gap requirement', () => {
    const state = makeState({
      gapNeeds: [
        [null, { kind: 'needed', design: 'SPADE', value: 7 }, { kind: 'anySuit', value: 8 }, { kind: 'blocked' }],
      ],
    });
    const text = formatGapsState(state);
    expect(text).toContain('[1] gap at 0,1 (needs ♠ 7)');
    expect(text).toContain('[2] gap at 0,2 (needs any suit 8)');
    expect(text).toContain('[3] gap at 0,3 (blocked)');
  });

  it('shows only available actions and concrete stalemate recovery', () => {
    const text = formatGapsState(
      makeState({ canUndo: false, redealsRemaining: 0, redealsUsed: 3, isStalemate: true, undoToEscape: 2 }),
    );
    expect(text).toContain('redeals: 3/3 (locked)');
    expect(text).toContain('with un 2');
    expect(text).not.toContain('Move a card to an empty cell with: p');
    expect(text).not.toContain('Use u to undo');
    expect(text).not.toContain('Use d to redeal');
  });

  it('prints no hint once and renders a requested hint as an executable indexed command', () => {
    expect(formatGapsState(makeState({ messageCode: 'gaps.noHint' }))).toContain('No hint available.');
    const hint = { fromRow: 1, fromCol: 2, toRow: 0, toCol: 12 };
    expect(formatGapsState(makeState({ hint, messageCode: 'gaps.hintAvailable' }))).toContain('command: p 15 0 12');
  });

  it('uses reset in ended phases and avoids duplicating clear messages', () => {
    const text = formatGapsState(makeState({ phase: 1, message: '日本語のクリア表示' }));
    expect(text).toContain('Use r to reset.');
    expect(text).not.toContain('日本語のクリア表示');
  });
});
