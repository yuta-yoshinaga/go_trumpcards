import type { Card, StHelenaResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const ASCENDING_COUNT = 4;

// Render a foundation pile's top card, or an empty slot.
function topOrEmpty(pile: Card[]): string {
  return pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]';
}

/** Format a StHelena Solitaire game state as terminal text. */
export function formatStHelenaState(state: StHelenaResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('StHelena'));

  const asc = state.foundation.slice(0, ASCENDING_COUNT).map(topOrEmpty);
  const desc = state.foundation.slice(ASCENDING_COUNT).map(topOrEmpty);
  lines.push(`asc:  ${asc.join(' | ')}`);
  lines.push(`desc: ${desc.join(' | ')}`);
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

  lines.push(`moves: ${state.moveCount} | redeals: ${state.redealsRemaining} | undo:${state.canUndo ? 'yes' : 'no'}`);

  if (state.hint && isRequestedHint(state)) {
    if (state.hint.redeal) {
      lines.push('HINT: redeal (d)');
    } else {
      const target = state.hint.toCol >= 0 ? `${state.hint.toZone}${state.hint.toCol}` : state.hint.toZone;
      lines.push(`HINT: t${state.hint.fromCol} → ${target}`);
    }
  }
  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
