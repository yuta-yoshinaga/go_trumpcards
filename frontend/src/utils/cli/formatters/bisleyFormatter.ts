import type { BisleyResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

/** Format a Bisley game state as terminal text. */
export function formatBisleyState(state: BisleyResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Bisley'));

  const topOf = (piles: BisleyResponse['aceFoundations']) =>
    piles.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]')).join(' | ');

  lines.push(`ascending  (A->K): ${topOf(state.aceFoundations)}`);
  lines.push(`descending (K->A): ${topOf(state.kingFoundations)}`);
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

  if (state.hint) {
    const target = state.hint.toZone === 'tableau' ? `t${state.hint.toIdx}` : `${state.hint.toZone}${state.hint.toIdx}`;
    lines.push(`HINT: t${state.hint.fromCol} → ${target}`);
  }
  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
