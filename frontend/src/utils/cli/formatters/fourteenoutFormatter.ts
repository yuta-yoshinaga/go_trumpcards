import type { FourteenOutResponse } from '../../../types/card';
import { FourteenOutPhase } from '../../../types/phases';
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
    `removed: ${state.removedCount}/52 | removable pairs: ${state.removablePairs} | undo available: ${state.canUndo ? 'yes' : 'no'}`,
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
    const pairs = Object.entries(state.legalTargets)
      .filter(([key]) => key.startsWith('tableau-'))
      .flatMap(([key, targets]) => targets.tableau.map((to) => `p ${key.slice('tableau-'.length)} ${to}`));
    lines.push(
      `Available commands: ${[...pairs, ...(state.canUndo ? ['u (undo)'] : []), 'h (hint)', 'g (give up)', 'r (reset)'].join(', ')}`,
    );
  } else {
    lines.push(
      state.phase === FourteenOutPhase.GAME_CLEAR ? 'All cards removed. Game clear!' : 'No more moves. Game over.',
    );
    lines.push('Available command: r (reset)');
  }
  if (state.messageCode === 'fourteenout.noHint') lines.push('No hint available.');
  if (state.hint && isRequestedHint(state))
    lines.push(`HINT: Remove columns ${state.hint.fromCol} and ${state.hint.toCol}.`);
  lines.push(formatSeparator());
  return lines.join('\n');
}
