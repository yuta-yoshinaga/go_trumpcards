import type { NapoleonsSquareResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Napoleon's Square game state as terminal text. */
export function formatNapoleonsSquareState(state: NapoleonsSquareResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader("Napoleon's Square"));

  const fnd = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`foundations: ${fnd.join(' | ')}`);

  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop} (${state.waste.length})`);
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

  if (state.hint && isRequestedHint(state)) {
    const from =
      state.hint.fromZone === 'tableau' ? `t${state.hint.fromCol}[${state.hint.cardIndex}]` : state.hint.fromZone;
    const to = state.hint.toCol >= 0 ? `${state.hint.toZone}${state.hint.toCol}` : state.hint.toZone;
    lines.push(`HINT: ${from} → ${to}`);
  }
  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
