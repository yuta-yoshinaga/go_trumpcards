import type { GapsResponse } from '../../../types/card';
import { GapsPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [GapsPhase.PLAYING]: 'Playing',
  [GapsPhase.GAME_CLEAR]: 'GameClear',
  [GapsPhase.GAME_OVER]: 'GameOver',
};

/** Format a Gaps game state as terminal text. */
export function formatGapsState(state: GapsResponse): string {
  const lines = [
    formatHeader('Gaps'),
    `phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
    `moves: ${state.moveCount}  redeals remaining: ${state.redealsRemaining}`,
    '',
  ];
  lines.push('Cards (p <idx> <row> <col> moves a card to an empty cell):');
  state.grid.forEach((row, rowIdx) => {
    row.forEach((card, colIdx) => {
      if (!card) return;
      const idx = rowIdx * 13 + colIdx;
      const locked = colIdx < (state.lockedPrefixLengths[rowIdx] ?? 0);
      lines.push(
        `[${idx}] ${locked ? 'x ' : ''}${formatCard(card)} at ${rowIdx},${colIdx}${locked ? ' (locked)' : ''}`,
      );
    });
  });

  if (state.phase === GapsPhase.PLAYING) {
    lines.push(
      '',
      'Move a card to an empty cell with: p <idx> <row> <col>',
      'Use u to undo, d to redeal, h for a hint, or g to give up.',
    );
    if (state.isStalemate) lines.push(`Stalemate. Undo ${state.undoToEscape ?? 0} moves to escape, or give up.`);
  }
  if (state.hint && isRequestedHint(state)) {
    lines.push(
      `HINT: move card at ${state.hint.fromRow},${state.hint.fromCol} to ${state.hint.toRow},${state.hint.toCol}`,
    );
  }
  if (state.message) lines.push(state.message);
  if (state.phase === GapsPhase.GAME_CLEAR) lines.push(`Game clear in ${state.moveCount} moves.`);
  if (state.phase === GapsPhase.GAME_OVER) lines.push('Game over.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
