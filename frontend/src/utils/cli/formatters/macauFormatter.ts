import type { MacauResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const SUIT_NAMES: Record<number, string> = { 1: 'Spade', 2: 'Clover', 3: 'Heart', 4: 'Diamond' };

/** Format a Macau game state as terminal text. */
export function formatMacauState(state: MacauResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Macau'));
  lines.push(
    `round: ${state.roundNumber}  draw pile: ${state.drawPileCount}  direction: ${state.direction < 0 ? '<-' : '->'}`,
  );
  if (state.discardTop) lines.push(`discard: ${formatCard(state.discardTop)}`);
  if (state.chosenSuit > 0) lines.push(`chosen suit: ${SUIT_NAMES[state.chosenSuit] ?? '?'}`);
  if (state.penaltyDrawCount > 0) lines.push(`draw penalty: ${state.penaltyDrawCount}`);
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
  if (state.phase === 2) lines.push('Declare "Macau!" (dc) or skip (sk)');

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
