import type { TriPeaksResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a TriPeaks game state as terminal text. */
export function formatTripeaksState(state: TriPeaksResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('TriPeaks'));
  lines.push(
    `stock: ${state.stockCount} | waste: ${state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]'}`,
  );
  lines.push('----------');

  for (let row = 0; row < state.layout.length; row++) {
    const cards = state.layout[row].map((tc) => {
      if (tc.removed) return '    ';
      if (!tc.exposed) return ' ?? ';
      return tc.card ? `${formatCard(tc.card).padEnd(4)}` : '    ';
    });
    lines.push(cards.join(' '));
  }
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}`);

  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.hint && isRequestedHint(state)) lines.push(`HINT: (${state.hint.row},${state.hint.col})`);
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
