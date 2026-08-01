import type { SpiderResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const DIFF_NAMES: Record<number, string> = { 1: '1 suit', 2: '2 suits', 4: '4 suits' };

/** Format a Spider Solitaire game state as terminal text. */
export function formatSpiderState(state: SpiderResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Spider Solitaire'));
  lines.push(
    `stock: ${state.stockCount} | completed: ${state.completedSuits}/8 | diff: ${DIFF_NAMES[state.difficulty] ?? state.difficulty}`,
  );
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
  if (state.hint && isRequestedHint(state)) {
    lines.push(`HINT: col${state.hint.fromCol}[${state.hint.cardIndex}] \u2192 col${state.hint.toCol}`);
  }
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
