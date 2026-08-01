import type { OsmosisResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format an Osmosis game state as terminal text. */
export function formatOsmosisState(state: OsmosisResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Osmosis'));

  // Foundation rows
  for (let i = 0; i < state.foundation.length; i++) {
    const pile = state.foundation[i];
    const top = pile.length > 0 ? `${formatCard(pile[pile.length - 1])} (${pile.length})` : '[  ]';
    lines.push(`foundation${i}: ${top}`);
  }
  lines.push(`base: ${state.baseRank}`);
  lines.push('----------');

  // Reserve columns
  for (let i = 0; i < state.reserve.length; i++) {
    const pile = state.reserve[i];
    const top = pile.length > 0 ? `${formatCard(pile[pile.length - 1])} (${pile.length})` : '[empty]';
    lines.push(`reserve${i}: ${top}`);
  }
  lines.push('----------');

  // Stock & waste
  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop}`);

  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);

  if (state.hint && isRequestedHint(state)) {
    const from = state.hint.fromZone === 'reserve' ? `reserve${state.hint.fromCol}` : 'waste';
    lines.push(`HINT: ${from} → foundation${state.hint.toCol}`);
  }
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
