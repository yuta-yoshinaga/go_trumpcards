import type { PyramidResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Pyramid Solitaire game state as terminal text. */
export function formatPyramidState(state: PyramidResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Pyramid'));
  lines.push(
    `stock: ${state.stockCount} | waste: ${state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]'}`,
  );
  lines.push('----------');

  // Pyramid rows
  for (let row = 0; row < state.pyramid.length; row++) {
    const padding = ' '.repeat((state.pyramid.length - 1 - row) * 2);
    const cards = state.pyramid[row].map((pc) => {
      if (pc.removed) return '    ';
      if (!pc.exposed) return ' ?? ';
      return pc.card ? formatCard(pc.card).padEnd(4) : '    ';
    });
    lines.push(`${padding}${cards.join(' ')}`);
  }
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}`);

  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.hint && isRequestedHint(state))
    lines.push(`HINT: (${state.hint.row1},${state.hint.col1}) + (${state.hint.row2},${state.hint.col2})`);
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
