import type { CrazyEightsResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const SUIT_NAMES: Record<number, string> = { 1: 'Spade', 2: 'Clover', 3: 'Heart', 4: 'Diamond' };

/** Format a Crazy Eights game state as terminal text. */
export function formatCrazyeightsState(state: CrazyEightsResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Crazy Eights'));
  lines.push(`round: ${state.roundNumber}  draw pile: ${state.drawPileCount}`);
  if (state.discardTop) lines.push(`discard: ${formatCard(state.discardTop)}`);
  if (state.chosenSuit > 0) lines.push(`chosen suit: ${SUIT_NAMES[state.chosenSuit] ?? '?'}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: total=${p.cumulativeScore} round=${p.roundScore} cards=${p.cardCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.phase === 1) lines.push('Choose a suit (suit <spade|clover|heart|diamond>)');

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
