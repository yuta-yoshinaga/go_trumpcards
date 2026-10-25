import type { GapsResponse } from '../../../types/card';
import { GapsPhase } from '../../../types/phases';
import { cardAlt } from '../../cardAlt';
import { valueName } from '../../cardUtils';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [GapsPhase.PLAYING]: 'Playing',
  [GapsPhase.GAME_CLEAR]: 'GameClear',
  [GapsPhase.GAME_OVER]: 'GameOver',
};

/** Format a Gaps game state as terminal text. */
export function formatGapsState(state: GapsResponse): string {
  const redealLimit = state.redealsUsed + state.redealsRemaining;
  const lines = [
    formatHeader('Gaps'),
    `phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
    `moves: ${state.moveCount}  redeals: ${state.redealsUsed}/${redealLimit}${state.redealsRemaining <= 0 ? ' (locked)' : ''}`,
    '',
    'Cards (p <idx> <row> <col> moves a card to an empty cell):',
  ];
  state.grid.forEach((row, rowIdx) => {
    lines.push(`Row ${rowIdx}:`);
    row.forEach((card, colIdx) => {
      const idx = rowIdx * 13 + colIdx;
      if (!card) {
        const gap = state.gapNeeds?.[rowIdx]?.[colIdx];
        const need =
          gap?.kind === 'needed'
            ? `needs ${cardAlt({ design: gap.design, value: gap.value })}`
            : gap?.kind === 'anySuit'
              ? `needs any suit ${valueName(gap.value)}`
              : gap?.kind === 'blocked'
                ? 'blocked'
                : 'open';
        lines.push(`  [${idx}] gap at ${rowIdx},${colIdx} (${need})`);
        return;
      }
      const locked = colIdx < (state.lockedPrefixLengths[rowIdx] ?? 0);
      lines.push(
        `[${idx}] ${locked ? 'x ' : ''}${formatCard(card)} at ${rowIdx},${colIdx}${locked ? ' (locked)' : ''}`,
      );
    });
  });

  if (state.phase === GapsPhase.PLAYING) {
    lines.push('');
    if (!state.isStalemate) lines.push('Move a card to an empty cell with: p <idx> <row> <col>');
    if (state.canUndo) lines.push('Use u to undo.');
    if (state.redealsRemaining > 0) lines.push('Use d to redeal.');
    lines.push('Use h for a hint, or g to give up.');
    if (state.isStalemate) {
      if (state.undoToEscape && state.undoToEscape > 0) {
        lines.push(
          `Stalemate. Undo ${state.undoToEscape} moves to escape with un ${state.undoToEscape}, or g to give up.`,
        );
      } else {
        lines.push(`Stalemate. Use ${state.canUndo ? 'u to undo' : 'r to reset'}, or g to give up.`);
      }
    }
  } else {
    lines.push('', 'Use r to reset.');
  }
  if (state.messageCode === 'gaps.noHint') lines.push('No hint available.');
  else if (state.hint && isRequestedHint(state)) {
    const fromIdx = state.hint.fromRow * 13 + state.hint.fromCol;
    lines.push(
      `HINT: move card at ${state.hint.fromRow},${state.hint.fromCol} to ${state.hint.toRow},${state.hint.toCol}; command: p ${fromIdx} ${state.hint.toRow} ${state.hint.toCol}`,
    );
  }
  if (state.message && state.phase !== GapsPhase.GAME_CLEAR) lines.push(state.message);
  if (state.phase === GapsPhase.GAME_CLEAR) lines.push(`Game clear in ${state.moveCount} moves.`);
  if (state.phase === GapsPhase.GAME_OVER) lines.push('Game over.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
