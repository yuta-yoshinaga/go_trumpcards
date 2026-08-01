import type { EightOffResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format an Eight Off game state as terminal text. */
export function formatEightoffState(state: EightOffResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Eight Off'));

  const cells = state.freeCells.map((c, i) => `[c${i}]${c ? formatCard(c) : '    '}`);
  lines.push(`cells: ${cells.join(' ')}`);

  const fnd = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`foundation: ${fnd.join(' | ')}`);
  lines.push('----------');

  for (let col = 0; col < state.tableau.length; col++) {
    const column = state.tableau[col];
    if (column.length === 0) {
      lines.push(`col${col}: [empty]`);
    } else {
      const cardStrs = column.map((c, i) => (c ? `[${i}]${formatCard(c)}` : '[?]'));
      lines.push(`col${col}: ${cardStrs.join(' ')}`);
    }
  }
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}`);

  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.hint && isRequestedHint(state)) {
    lines.push(
      `HINT: ${state.hint.fromZone}${state.hint.fromCol >= 0 ? state.hint.fromCol : ''} → ${state.hint.toZone}${state.hint.toCol >= 0 ? state.hint.toCol : ''}`,
    );
  }
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
