import type { MatrimonyResponse } from '../../../types/games/matrimony';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Matrimony game state as terminal text. */
export function formatMatrimonyState(state: MatrimonyResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Matrimony'));

  const fnd = state.foundation.map((pile, i) => {
    const start = i < 2 ? 'Q♠↓' : 'J♦↑';
    return `${start}:${pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'}`;
  });
  lines.push(`foundations: ${fnd.join(' | ')}`);

  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop}`);
  lines.push('----------');

  // A slot holds exactly one card; four to a row mirrors the board.
  for (let row = 0; row < state.tableau.length; row += 4) {
    const cells = state.tableau.slice(row, row + 4).map((card, i) => {
      const slot = row + i;
      return `[${slot}]${card ? formatCard(card) : ' .. '}`;
    });
    lines.push(cells.join(' '));
  }
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);

  if (state.hint && isRequestedHint(state)) {
    const from = state.hint.fromIdx >= 0 ? `t${state.hint.fromIdx}` : state.hint.fromZone;
    const to = state.hint.toIdx >= 0 ? `${state.hint.toZone}${state.hint.toIdx}` : state.hint.toZone;
    lines.push(`HINT: ${from} → ${to}`);
  }
  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
