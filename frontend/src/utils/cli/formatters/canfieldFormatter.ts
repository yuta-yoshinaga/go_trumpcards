import type { CanfieldResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Canfield game state as terminal text. */
export function formatCanfieldState(state: CanfieldResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Canfield'));

  // Stock & waste
  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop}`);

  // Reserve
  const reserveTop = state.reserve.length > 0 ? formatCard(state.reserve[state.reserve.length - 1]) : '[  ]';
  lines.push(`reserve: ${reserveTop} (${state.reserve.length})`);

  // Foundation
  const fnd = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`foundation: ${fnd.join(' | ')}  base: ${state.baseRank}`);
  lines.push('----------');

  // Tableau columns
  for (let col = 0; col < state.tableau.length; col++) {
    const column = state.tableau[col];
    if (column.length === 0) {
      lines.push(`t${col}: [empty]`);
      continue;
    }
    const cardStrs = column.map((c, i) => `[${i}]${formatCard(c.card)}`);
    lines.push(`t${col}: ${cardStrs.join(' ')}`);
  }
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);

  if (state.hint && isRequestedHint(state)) {
    const fromCol = state.hint.fromCol >= 0 ? state.hint.fromCol : '';
    const toCol = state.hint.toCol >= 0 ? state.hint.toCol : '';
    lines.push(`HINT: ${state.hint.fromZone}${fromCol}[${state.hint.cardIndex}] → ${state.hint.toZone}${toCol}`);
  }
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
