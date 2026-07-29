import type { AmericanToadResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

/** Format an American Toad game state as terminal text. */
export function formatAmericanToadState(state: AmericanToadResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('American Toad'));
  lines.push(`base rank: ${state.baseRank}`);

  const fnd = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`foundations: ${fnd.join(' | ')}`);

  // Once the reserve is gone the empty-column rule changes, so its depth is
  // status information, not decoration.
  const reserveTop = state.reserve.length > 0 ? formatCard(state.reserve[state.reserve.length - 1]) : '[  ]';
  lines.push(`reserve: ${reserveTop} (${state.reserve.length})`);

  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop}`);
  lines.push('----------');

  for (let col = 0; col < state.tableau.length; col++) {
    const column = state.tableau[col];
    if (column.length === 0) {
      lines.push(`t${col}: [empty]`);
      continue;
    }
    const cardStrs = column.map((c, i) => (c.card ? `[${i}]${formatCard(c.card)}` : '[?]'));
    lines.push(`t${col}: ${cardStrs.join(' ')}`);
  }
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);
  if (state.canRedeal) lines.push('Stock is out - d redeals the waste (once only)');

  if (state.hint) {
    const from =
      state.hint.fromZone === 'tableau' ? `t${state.hint.fromIdx}[${state.hint.cardIndex}]` : state.hint.fromZone;
    const to = state.hint.toIdx >= 0 ? `${state.hint.toZone}${state.hint.toIdx}` : state.hint.toZone;
    lines.push(`HINT: ${from} → ${to}`);
  }
  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
