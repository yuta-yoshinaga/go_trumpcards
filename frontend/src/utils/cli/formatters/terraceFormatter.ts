import type { TerraceResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Terrace game state as terminal text. */
export function formatTerraceState(state: TerraceResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Terrace'));
  // Until the base rank is fixed that one decision dominates, so it leads.
  lines.push(state.awaitingBaseRank ? 'base rank: not set yet' : `base rank: ${state.baseRank}`);

  const fnd = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`foundations: ${fnd.join(' | ')}`);

  // The terrace never refills, so its depth is the number that matters.
  const terraceTop = state.reserve.length > 0 ? formatCard(state.reserve[state.reserve.length - 1]) : '[  ]';
  lines.push(`terrace: ${terraceTop} (${state.reserve.length}, foundations only)`);

  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop}`);
  lines.push('----------');

  for (let pile = 0; pile < state.tableau.length; pile++) {
    const cards = state.tableau[pile];
    if (cards.length === 0) {
      lines.push(`t${pile}: [empty]`);
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
