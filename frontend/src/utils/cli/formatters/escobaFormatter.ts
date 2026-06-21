import type { EscobaResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName } from '../formatterBase';

/** Format an Escoba game state as terminal text (4 players, free-for-all). */
export function formatEscobaState(state: EscobaResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Escoba'));
  lines.push(`Phase: ${state.phase} / Turn: ${state.gameEndFlag ? 'End' : `Player ${state.currentTurn}`}`);
  lines.push('----------');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: hand=${p.handCount} captured=${p.capturedCount} escobas=${p.escobaCount} score=${p.score}`);
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

  lines.push(`Stock: ${state.stockRemaining}`);

  if (state.message) lines.push(state.message);

  return lines.join('\n');
}
