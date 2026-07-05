import type { IndianRummyResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'DRAW',
  1: 'DISCARD',
  2: 'ROUND END',
  3: 'GAME END',
};

/** Format an Indian Rummy game state as terminal text. */
export function formatIndianrummyState(state: IndianRummyResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Indian Rummy'));
  lines.push(`round: ${state.roundNumber}/${state.targetRounds}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`wild joker: ${state.wildJoker ? formatCard(state.wildJoker) : '[  ]'}`);
  lines.push(`discard: ${state.discardTop ? formatCard(state.discardTop) : '[  ]'} | stock: ${state.drawPileCount}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: total=${p.cumulativeScore} round=${p.roundScore} cards=${p.cardCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winner}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
