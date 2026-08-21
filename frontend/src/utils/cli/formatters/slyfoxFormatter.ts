import type { SlyFoxResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

/** Format a Sly Fox game state as terminal text. */
export function formatSlyFoxState(state: SlyFoxResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Sly Fox'));

  // **周の進みは盤から読めない。**書かないと、リザーブが送れない理由が分からない。
  lines.push(`stock: ${state.stockCount}`);
  lines.push(
    state.reserveLocked
      ? `this round: ${state.dealtThisCycle}/${state.dealCycle} dealt - ${state.dealCycle - state.dealtThisCycle} more before the reserve opens`
      : 'the reserve is open',
  );

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
    lines.push(`slot${col}: ${top} (${pile.length})`);
  }

  lines.push(`moves: ${state.moveCount}`);
  if (state.message) lines.push(state.message);
  return lines.join('\n');
}
