import type { CurdsAndWheyResponse } from '../../../types/card';
import { CurdsAndWheyPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES: Readonly<Record<number, string>> = {
  [CurdsAndWheyPhase.PLAYING]: 'Playing',
  [CurdsAndWheyPhase.GAME_CLEAR]: 'Game Clear',
  [CurdsAndWheyPhase.GAME_OVER]: 'Game Over',
};

/** Format a Curds and Whey response for terminal play. */
export function formatCurdsAndWheyState(state: CurdsAndWheyResponse): string {
  const lines = [formatHeader('Curds and Whey'), `phase: ${PHASE_NAMES[state.phase] ?? state.phase}`];
  lines.push(`completed suits: ${state.completedSuits}/4 | moves: ${state.moveCount}`);
  for (const [col, cards] of state.columns.entries()) {
    const starts = state.runStarts[col] ?? [];
    lines.push(
      `Column ${col}: ${cards.map((card, idx) => `[${idx}]${starts.includes(idx) ? '*' : ''}${formatCard(card)}`).join('  ') || '[empty]'}`,
    );
  }
  if (state.phase === CurdsAndWheyPhase.PLAYING) {
    const suggested = state.runStarts.flatMap((indices, fromCol) =>
      indices.flatMap((idx) =>
        (state.legalTargets[`tableau-${fromCol}-${idx}`]?.tableau ?? []).map((toCol) => `p ${fromCol} ${idx} ${toCol}`),
      ),
    );
    lines.push('Move a run with p <fromCol> <cardIndex> <toCol>; card indices start at 0.');
    lines.push(
      `Available commands: ${[...suggested, ...(state.canUndo ? ['u (undo)'] : []), 'h (hint)', 'g (give up)', 'r (reset)'].join(', ')}`,
    );
  } else {
    if (state.phase === CurdsAndWheyPhase.GAME_CLEAR) lines.push('All suits completed. Game clear!');
    if (state.phase === CurdsAndWheyPhase.GAME_OVER) lines.push('No moves remain. Game over.');
    lines.push('Available command: r (reset)');
  }
  if (state.messageCode === 'curdsandwhey.noHint') lines.push('No hint available.');
  if (state.hint && isRequestedHint(state)) {
    const card = state.columns[state.hint.fromCol]?.[state.hint.cardIndex];
    lines.push(
      `HINT: Move ${card ? formatCard(card) : `card [${state.hint.cardIndex}]`} from column ${state.hint.fromCol} to column ${state.hint.toCol}.`,
    );
  }
  if (state.message && state.phase === CurdsAndWheyPhase.PLAYING && state.messageCode !== 'curdsandwhey.noHint')
    lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
