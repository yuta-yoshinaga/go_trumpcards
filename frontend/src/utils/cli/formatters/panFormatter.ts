import type { PanResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'DRAW',
  1: 'PLAY',
  2: 'ROUND END',
  3: 'GAME END',
};

/** Format a Panguingue (Pan) game state as terminal text. */
export function formatPanState(state: PanResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Panguingue / Pan'));
  lines.push(`round: ${state.roundNumber}/${state.targetRounds}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`discard: ${state.discardTop ? formatCard(state.discardTop) : '[  ]'} | stock: ${state.drawPileCount}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: total=${p.cumulativeScore} chips=${p.chips} melded=${p.meldedCount} cards=${p.cardCount}`);
    p.laidMelds.forEach((meld, idx) => {
      lines.push(`  meld[${idx}]: ${meld.cards.map(formatCard).join(' ')}`);
    });
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
