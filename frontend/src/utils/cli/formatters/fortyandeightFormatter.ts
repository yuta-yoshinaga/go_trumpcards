import type { FortyAndEightResponse } from '../../../types/card';
import { FortyAndEightPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Forty and Eight game state as terminal text. */
export function formatFortyandeightState(state: FortyAndEightResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Forty and Eight'));

  // Foundations (8 piles)
  const fnd = state.foundation.map((pile) => (pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]'));
  lines.push(`foundation: ${fnd.join(' | ')}`);

  // Stock and waste
  const wasteTop = state.waste.length > 0 ? formatCard(state.waste[state.waste.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockCount}  waste: ${wasteTop}  redeal:${state.redealUsed ? 'used' : 'available'}`);
  lines.push('----------');

  // Tableau columns
  for (let col = 0; col < state.tableau.length; col++) {
    const column = state.tableau[col];
    if (column.length === 0) {
      lines.push(`t${col}: [empty]`);
      continue;
    }
    const cardStrs = column.map((c, i) => (c.faceUp && c.card ? `[${i}]${formatCard(c.card)}` : '[?]'));
    lines.push(`t${col}: ${cardStrs.join(' ')}`);
  }
  lines.push('----------');

  lines.push(`moves: ${state.moveCount}  undo:${state.canUndo ? 'yes' : 'no'}`);

  if (state.hint && isRequestedHint(state)) {
    const from = state.hint.fromZone === 'waste' ? 'waste' : `t${state.hint.fromCol}[${state.hint.cardIndex}]`;
    const target = state.hint.toCol >= 0 ? `${state.hint.toZone}${state.hint.toCol}` : state.hint.toZone;
    lines.push(`HINT: ${from} → ${target}`);
  }
  if (state.isStalemate) lines.push('Stalemate - no more moves possible');
  if (state.message) lines.push(state.message);
  if (state.phase === FortyAndEightPhase.GAME_CLEAR) lines.push('Congratulations! You win!');

  lines.push(formatSeparator());
  return lines.join('\n');
}
