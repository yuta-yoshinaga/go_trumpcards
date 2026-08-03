import type { FreeCellResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a FreeCell game state as terminal text. */
export function formatFreecellState(state: FreeCellResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('FreeCell'));

  // Free cells
  const cells = state.freeCells.map((c, i) => `[c${i}]${c ? formatCard(c) : '    '}`);
  lines.push(`cells: ${cells.join(' ')}`);

  // Foundation
  const fnd = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`foundation: ${fnd.join(' | ')}`);
  lines.push('----------');

  // Tableau
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
      `HINT: ${state.hint.fromZone}${state.hint.fromCol >= 0 ? state.hint.fromCol : ''} \u2192 ${state.hint.toZone}${state.hint.toCol >= 0 ? state.hint.toCol : ''}`,
    );
  }
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
