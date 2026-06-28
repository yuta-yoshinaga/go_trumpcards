import type { PrsiResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

/** Format a Prší game state as terminal text. */
export function formatPrsiState(state: PrsiResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Prší'));
  lines.push(`draw pile: ${state.drawPileCount}`);
  if (state.discardTop) lines.push(`discard: ${formatCard(state.discardTop)}`);
  if (state.penaltyDrawCount > 0) lines.push(`penalty: ${state.penaltyDrawCount} (stack a 7 or draw)`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: cards=${p.cardCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentPlayerIdx, state.players[state.currentPlayerIdx]?.isHuman ?? false);
    lines.push(`turn: ${current}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winner}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
