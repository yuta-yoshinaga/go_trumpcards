import type { ColoradoResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

/** Format a Colorado game state as terminal text. */
export function formatColoradoState(state: ColoradoResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Colorado'));

  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop}`);

  // Half the foundations build up and half build down, and the board cannot be
  // read without knowing which is which -- so each one carries its arrow.
  const fnd = state.foundation.map((pile, i) => {
    const dir = state.foundationAscending[i] ? '↑' : '↓';
    return `${dir}${pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'}`;
  });
  lines.push(`foundation: ${fnd.join(' | ')}`);
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
