import type { SpideretteResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

/** Format a Spiderette game state as terminal text. */
export function formatSpideretteState(state: SpideretteResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Spiderette'));
  lines.push(`stock: ${state.stockCount} | completed: ${state.completedSuits}/4`);
  lines.push('----------');

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

  lines.push(`moves: ${state.moveCount} | score: ${state.score}`);

  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.hint) {
    lines.push(`HINT: col${state.hint.fromCol}[${state.hint.cardIndex}] → col${state.hint.toCol}`);
  }
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
