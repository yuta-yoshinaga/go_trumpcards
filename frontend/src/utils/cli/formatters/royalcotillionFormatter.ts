import type { RoyalCotillionResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a RoyalCotillion game state as terminal text. */
export function formatRoyalCotillionState(state: RoyalCotillionResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Royal Cotillion'));

  // Half the piles start at the Ace and half at the deuce, and both wrap, so
  // the series marker is the only way to read what a pile wants next.
  const fnd = state.foundation.map((pile, i) => {
    const series = state.foundationOdd?.[i] ? 'A:' : '2:';
    return `${series}${pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'}`;
  });
  lines.push(`foundations: ${fnd.join(' | ')}`);

  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop}`);
  lines.push('----------');

  // A slot holds exactly one card; four to a row mirrors the board.
  for (let row = 0; row < state.tableau.length; row += 4) {
    const cells = state.tableau.slice(row, row + 4).map((card, i) => {
      const slot = row + i;
      return `[${slot}]${card ? formatCard(card) : ' .. '}`;
    });
    lines.push(cells.join(' '));
  }
  lines.push('----------');

  // An emptied reserve pile is never refilled, so say so rather than print a
  // bare blank that reads like a rendering gap.
  for (let pile = 0; pile < state.reserve.length; pile++) {
    const cards = state.reserve[pile];
    if (cards.length === 0) {
      lines.push(`r${pile}: [empty] (never refilled)`);
      continue;
    }
    lines.push(`r${pile}: ${cards.map((c, i) => `[${i}]${formatCard(c)}`).join(' ')}`);
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
