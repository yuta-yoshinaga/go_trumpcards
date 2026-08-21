import type { BigBenResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Big Ben game state as terminal text. */
export function formatBigBenState(state: BigBenResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Big Ben'));

  // **山札の残りを出す。**補充がこのゲームの逃げ道なので、あと何枚あるかが
  // 読めないと「詰んだのか、配れば動くのか」が分からない。
  lines.push(`stock: ${state.stockCount}`);

  // One line per face: the target rank has to sit next to the pile it applies to.
  for (let i = 0; i < state.foundation.length; i++) {
    const f = state.foundation[i];
    const top = f.cards.length > 0 ? formatCard(f.cards[f.cards.length - 1]) : '[  ]';
    lines.push(`f${i}: ${top} -> ${f.targetRank}${f.complete ? ' done' : ''}`);
  }
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
    const to = state.hint.toZone === 'foundation' ? `f${state.hint.toIdx}` : `t${state.hint.toIdx}`;
    lines.push(`HINT: t${state.hint.fromCol} → ${to}`);
  }
  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
