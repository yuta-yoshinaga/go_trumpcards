import type { BraidResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Render one row of single-card slots, keeping empty slots in place. */
function formatSlots(label: string, slots: (Card | null)[]): string {
  const cells = slots.map((c, i) => `${label}${i}:${c ? formatCard(c) : '[  ]'}`);
  return cells.join(' ');
}

type Card = Parameters<typeof formatCard>[0];

/** Format a Braid game state as terminal text. */
export function formatBraidState(state: BraidResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Braid'));
  // Until the direction is fixed that one decision dominates, so it leads.
  const dir = state.direction === 1 ? 'up' : 'down';
  lines.push(
    state.awaitingDirection
      ? `base rank: ${state.baseRank} (direction not set yet - use dir a / dir d)`
      : `base rank: ${state.baseRank} (${dir}, in suit)`,
  );

  const fnd = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`foundations: ${fnd.join(' | ')}`);

  // Only the tail is available, and the braid only shrinks, so its depth is the
  // number that matters.
  const braidTail = state.braid.length > 0 ? formatCard(state.braid[state.braid.length - 1]) : '[  ]';
  lines.push(`braid: ${braidTail} (${state.braid.length}, tail only)`);

  lines.push(formatSlots('fd', state.fields));
  lines.push(formatSlots('hp', state.helpers));

  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount} (${state.redealsLeft} redeal(s) left)  waste: ${wasteTop}`);
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);

  if (state.hint && isRequestedHint(state)) {
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
