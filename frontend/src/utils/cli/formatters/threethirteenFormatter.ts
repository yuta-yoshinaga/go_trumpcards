import type { ThreeThirteenResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'DRAW',
  1: 'DISCARD',
  2: 'ROUND END',
  3: 'GAME END',
};

/** Format a Three Thirteen game state as terminal text. */
export function formatThreeThirteenState(state: ThreeThirteenResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Three Thirteen'));
  lines.push(`round: ${state.round}/11  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`wild: ${state.wildRank}  deal: ${state.dealCount}`);
  lines.push(`discard: ${state.discardTop ? formatCard(state.discardTop) : '[  ]'} | stock: ${state.drawPileCount}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: total=${p.cumulativeScore} round=${p.roundScore} dead=${p.deadwood} cards=${p.cardCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.knockerIdx >= 0) {
    const knocker = formatPlayerName(state.knockerIdx, state.players[state.knockerIdx]?.isHuman ?? false);
    lines.push(`${knocker} knocked!`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winner}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
