import type { WhiteheadResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Whitehead game state as terminal text. */
export function formatWhiteheadState(state: WhiteheadResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Whitehead'));

  // Foundation
  const fndParts = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`Foundation: ${fndParts.join(' | ')}`);

  // Stock / Waste
  const wasteCard = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`Stock: ${state.stockCount} | Waste: ${wasteCard} | Draw: ${state.drawCount}`);
  lines.push('----------');

  // Tableau columns
  for (let col = 0; col < state.tableau.length; col++) {
    const column = state.tableau[col];
    if (column.length === 0) {
      lines.push(`col${col}: [empty]`);
    } else {
      const cardStrs = column.map((tc, i) => (tc.faceUp && tc.card ? `[${i}]${formatCard(tc.card)}` : '[?]'));
      lines.push(`col${col}: ${cardStrs.join(' ')}`);
    }
  }
  lines.push('----------');

  // Game info
  lines.push(`moves: ${state.moveCount} | score: ${state.score}`);

  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.hint && isRequestedHint(state)) {
    lines.push(
      `HINT: ${state.hint.fromZone}${state.hint.fromCol >= 0 ? state.hint.fromCol : ''} \u2192 ${state.hint.toZone}${state.hint.toCol >= 0 ? state.hint.toCol : ''}`,
    );
  }
  if (state.message) lines.push(state.message);

  // Win condition
  if (state.phase === 2) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
