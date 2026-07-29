import type { WindmillResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const CENTER_TARGET = 52;
const CORNER_TARGET = 13;

/** Format a Windmill game state as terminal text. */
export function formatWindmillState(state: WindmillResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Windmill'));

  const centerTop = state.center.length > 0 ? formatCard(state.center[state.center.length - 1]) : '[  ]';
  lines.push(`center: ${centerTop} (${state.center.length}/${CENTER_TARGET})`);

  const corners = state.corners.map((pile, i) =>
    pile.length > 0 ? `k${i}:${formatCard(pile[pile.length - 1])}(${pile.length}/${CORNER_TARGET})` : `k${i}:[  ]`,
  );
  lines.push(`corners: ${corners.join(' ')}`);

  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop}`);
  lines.push('----------');

  for (let i = 0; i < state.sails.length; i++) {
    const card = state.sails[i];
    lines.push(`s${i}: ${card ? formatCard(card) : '[empty]'}`);
  }
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);
  // The block is invisible in the layout, so it has to be stated.
  if (state.transferBlocked) lines.push('The next centre card must come from a sail or the waste');

  if (state.hint) {
    const from = state.hint.fromIdx >= 0 ? `${state.hint.fromZone}${state.hint.fromIdx}` : state.hint.fromZone;
    const to = state.hint.toIdx >= 0 ? `${state.hint.toZone}${state.hint.toIdx}` : state.hint.toZone;
    lines.push(`HINT: ${from} → ${to}`);
  }
  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
