import type { FourteenOutResponse } from '../../../types/card';
import { FourteenOutPhase } from '../../../types/phases';
import { fourteenOutCanUndo, fourteenOutRemovablePairs } from '../../fourteenOutRules';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES: Readonly<Record<number, string>> = {
  [FourteenOutPhase.PLAYING]: 'Playing',
  [FourteenOutPhase.GAME_CLEAR]: 'Game Clear',
  [FourteenOutPhase.GAME_OVER]: 'Game Over',
};

/** Format a Fourteen Out state for terminal play. */
export function formatFourteenOutState(state: FourteenOutResponse): string {
  const lines = [formatHeader('Fourteen Out'), `phase: ${PHASE_NAMES[state.phase] ?? state.phase}`];
  lines.push(
    `removed: ${state.removedCount}/52 | removable pairs: ${fourteenOutRemovablePairs(state).length} | undo available: ${fourteenOutCanUndo(state) ? 'yes' : 'no'}`,
  );
  lines.push('Columns (only the top card is playable; use p <from> <to>):');
  state.columns.forEach((column, index) => {
    const cards = column
      .map(
        (cell, cardIndex) =>
          `${cardIndex === column.length - 1 ? '*' : 'x'}${cell.card ? formatCard(cell.card) : '[empty]'}`,
      )
      .join(' ');
    lines.push(`[${index}] ${cards || '[empty]'}`);
  });
  if (state.phase === FourteenOutPhase.PLAYING) {
    const pairs = fourteenOutRemovablePairs(state).map(([from, to]) => `p ${from} ${to}`);
    lines.push(
      `Available commands: ${[...pairs, ...(state.canUndo ? ['u (undo)'] : []), 'h (hint)', 'g (give up)', 'r (reset)'].join(', ')}`,
    );
  } else {
    lines.push(
      state.phase === FourteenOutPhase.GAME_CLEAR ? 'All cards removed. Game clear!' : 'No more moves. Game over.',
    );
    lines.push('Available command: r (reset)');
  }
  if (state.phase === FourteenOutPhase.PLAYING && state.isStalemate) {
    lines.push('No more pairs. Undo, give up or reset.');
  }
  if (state.phase === FourteenOutPhase.PLAYING && state.message && !state.messageCode) lines.push(state.message);
  if (state.messageCode === 'fourteenout.noHint') lines.push('No hint available.');
  if (state.hint && isRequestedHint(state))
    lines.push(`HINT: Remove columns ${state.hint.fromCol} and ${state.hint.toCol}.`);
  lines.push(formatSeparator());
  return lines.join('\n');
}
