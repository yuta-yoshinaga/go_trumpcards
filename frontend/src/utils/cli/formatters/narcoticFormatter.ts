import type { NarcoticResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Narcotic game state as terminal text. */
export function formatNarcoticState(state: NarcoticResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Narcotic'));
  lines.push(`stock: ${state.stockCount} | discard: ${state.discardCount} | redeals: ${state.redealCount}`);
  lines.push('----------');

  state.columns.forEach((col, i) => {
    if (col.length === 0) {
      lines.push(`[${i}] (empty)`);
      return;
    }
    const cards = col.map((c, idx) => {
      const text = formatCard(c.card);
      return idx === col.length - 1 ? `[${text}]` : text;
    });
    lines.push(`[${i}] ${cards.join(' ')}`);
  });
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}`);

  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.hint && isRequestedHint(state)) {
    lines.push(state.hint.type === 'draw' ? 'HINT: deal' : `HINT: ${state.hint.type} col ${state.hint.col}`);
  }
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
