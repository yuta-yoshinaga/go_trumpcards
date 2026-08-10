import type { FourSeasonsResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

/** Format a Four Seasons game state as terminal text. */
export function formatFourSeasonsState(state: FourSeasonsResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Four Seasons'));

  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop}`);

  // The base rank is printed with the foundations because every corner builds
  // from it — reading the corners without it tells you nothing.
  const fnd = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`foundation: ${fnd.join(' | ')}  base: ${state.baseRank}`);
  lines.push('----------');

  for (let col = 0; col < state.tableau.length; col++) {
    const pile = state.tableau[col];
    const top = pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]';
    lines.push(`T${col}: ${top} (${pile.length})`);
  }

  lines.push(`moves: ${state.moveCount}`);
  if (state.message) lines.push(state.message);
  return lines.join('\n');
}
