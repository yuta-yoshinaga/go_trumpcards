import type { DiplomatResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Diplomat game state as terminal text. */
export function formatDiplomatState(state: DiplomatResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Diplomat'));

  const fnd = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`foundations: ${fnd.join(' | ')}`);

  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop}`);
  lines.push('----------');

  for (let pile = 0; pile < state.tableau.length; pile++) {
    const cards = state.tableau[pile];
    if (cards.length === 0) {
      // An empty column takes another column's top or the waste card -- never
      // the stock directly. Congress is the other way round.
      lines.push(`t${pile}: [empty] (tableau or waste only)`);
      continue;
    }
    lines.push(`t${pile}: ${cards.map((c, i) => `[${i}]${formatCard(c)}`).join(' ')}`);
  }
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);

  if (state.hint && isRequestedHint(state)) {
    const from = state.hint.fromIdx >= 0 ? `t${state.hint.fromIdx}` : state.hint.fromZone;
    const to = state.hint.toIdx >= 0 ? `${state.hint.toZone}${state.hint.toIdx}` : state.hint.toZone;
    lines.push(`HINT: ${from} → ${to}`);
  }
  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
