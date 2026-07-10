import type { MachiavelliResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'TURN',
  1: 'ROUND END',
  2: 'GAME END',
};

/** Format a Machiavelli game state as terminal text. */
export function formatMachiavelliState(state: MachiavelliResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Machiavelli'));
  lines.push(`round: ${state.roundNumber}/${state.targetRounds}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`stock: ${state.drawPileCount}`);
  lines.push('');

  lines.push('table melds:');
  if (state.table.length === 0) {
    lines.push('  (no melds yet)');
  } else {
    state.table.forEach((meld, idx) => {
      lines.push(`  [${idx}] ${meld.cards.map((c) => formatCard(c)).join(' ')}`);
    });
  }
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
