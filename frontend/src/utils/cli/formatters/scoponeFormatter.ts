import type { ScoponeResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName } from '../formatterBase';

/** Format a Scopone game state as terminal text (4 players / 2 teams). */
export function formatScoponeState(state: ScoponeResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Scopone'));
  lines.push(`Phase: ${state.phase} / Turn: ${state.gameEndFlag ? 'End' : `Player ${state.currentTurn}`}`);
  lines.push('----------');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name} team${p.team}: hand=${p.handCount} captured=${p.capturedCount} scopas=${p.scopaCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.tableCards.length > 0) {
    lines.push(`Table: ${formatIndexedCards(state.tableCards)}`);
  } else {
    lines.push('Table: (empty)');
  }

  lines.push(state.teamScores.map((sc, t) => `Team${t}: ${sc}`).join('  '));

  if (state.message) lines.push(state.message);

  return lines.join('\n');
}
