import type { BristolResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Bristol game state as terminal text. */
export function formatBristolState(state: BristolResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Bristol'));

  // Foundations
  for (let i = 0; i < state.foundation.length; i++) {
    const pile = state.foundation[i];
    const top = pile.length > 0 ? `${formatCard(pile[pile.length - 1])} (${pile.length})` : '[  ]';
    lines.push(`foundation${i}: ${top}`);
  }
  lines.push('----------');

  // Tableau columns (full column)
  for (let i = 0; i < state.tableau.length; i++) {
    const col = state.tableau[i];
    const cards = col.length > 0 ? col.map(formatCard).join(' ') : '[empty]';
    lines.push(`tableau${i}: ${cards}`);
  }
  lines.push('----------');

  // Fans (top card)
  for (let i = 0; i < state.fan.length; i++) {
    const pile = state.fan[i];
    const top = pile.length > 0 ? `${formatCard(pile[pile.length - 1])} (${pile.length})` : '[empty]';
    lines.push(`fan${i}: ${top}`);
  }
  lines.push('----------');

  lines.push(`stock: ${state.stockCount}`);
  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);

  if (state.hint && isRequestedHint(state)) {
    const from = state.hint.fromZone === 'fan' ? `fan${state.hint.fromCol}` : `tableau${state.hint.fromCol}`;
    const to = state.hint.toZone === 'foundation' ? `foundation${state.hint.toCol}` : `tableau${state.hint.toCol}`;
    lines.push(`HINT: ${from} → ${to}`);
  }
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
