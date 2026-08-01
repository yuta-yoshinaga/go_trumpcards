import type { SultanResponse } from '../../../types/card';
import { SultanPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Sultan of Turkey game state as terminal text. */
export function formatSultanState(state: SultanResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Sultan of Turkey'));

  // Foundations (8 King-based piles)
  const fnd = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`foundation: ${fnd.join(' | ')}`);

  // Divan reserve (8 slots; null = played/empty)
  const div = state.divan.map((c, i) => (c ? `[${i}]${formatCard(c)}` : `[${i}][  ]`));
  lines.push(`divan: ${div.join(' ')}`);

  // Stock and waste
  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop}  redeal:${state.canRedeal ? 'available' : 'used'}`);
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);

  if (state.hint && isRequestedHint(state)) {
    const from = state.hint.fromZone === 'waste' ? 'waste' : `divan[${state.hint.fromIdx}]`;
    lines.push(`HINT: ${from} → foundation${state.hint.toFoundation}`);
  }
  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.message) lines.push(state.message);
  if (state.phase === SultanPhase.GAME_CLEAR) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
