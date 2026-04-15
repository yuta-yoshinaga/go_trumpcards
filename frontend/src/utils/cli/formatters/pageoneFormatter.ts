import type { PageOneResponse } from '../../../types/card';
import { PageOnePhase } from '../../../types/phases';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

/** Format a Page One game state as terminal text. */
export function formatPageoneState(state: PageOneResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Page One'));
  lines.push(`round: ${state.roundNumber}  draw pile: ${state.drawPileCount}`);
  if (state.discardTop) lines.push(`discard: ${formatCard(state.discardTop)}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const declared = p.hasDeclared ? ' [PAGE ONE!]' : '';
    lines.push(`${name}: total=${p.cumulativeScore} round=${p.roundScore} cards=${p.cardCount}${declared}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.phase === PageOnePhase.MUST_DECLARE) lines.push('Declare Page One (dc) or skip (sk) to take penalty.');

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
