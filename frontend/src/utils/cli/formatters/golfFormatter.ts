import type { GolfResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Golf Solitaire game state as terminal text. */
export function formatGolfState(state: GolfResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Golf'));
  lines.push(
    `stock: ${state.stockCount} | waste: ${state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]'}`,
  );
  lines.push('----------');

  // Find max column height
  const maxHeight = Math.max(...state.layout.map((col) => col.length));
  for (let row = 0; row < maxHeight; row++) {
    const cards = state.layout.map((col) => {
      if (row >= col.length) return '    ';
      const gc = col[row];
      if (gc.removed) return '    ';
      if (!gc.exposed) return ' ?? ';
      return gc.card ? formatCard(gc.card).padEnd(4) : '    ';
    });
    lines.push(cards.join(' '));
  }
  // Column indices
  lines.push(state.layout.map((_, i) => `[${i}] `).join(' '));
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}`);

  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.hint && isRequestedHint(state)) lines.push(`HINT: col ${state.hint.col}`);
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
