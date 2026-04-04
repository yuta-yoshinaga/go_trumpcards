import type { GoFishResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const VALUE_NAMES: Record<number, string> = { 1: 'A', 11: 'J', 12: 'Q', 13: 'K' };

function rankName(r: number): string {
  return VALUE_NAMES[r] ?? String(r);
}

/** Format a Go Fish game state as terminal text. */
export function formatGofishState(state: GoFishResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Go Fish'));
  lines.push(`deck: ${state.deckRemaining}  turn: #${state.turnNumber}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const books = p.books.map((b) => rankName(b.rank)).join(', ');
    lines.push(`${name}: ${p.cardCount} cards, ${p.bookCount} books${books ? ` (${books})` : ''}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.lastAsk) {
    const asker = formatPlayerName(state.lastAsk.playerIdx, state.players[state.lastAsk.playerIdx]?.isHuman ?? false);
    const target = formatPlayerName(state.lastAsk.targetIdx, state.players[state.lastAsk.targetIdx]?.isHuman ?? false);
    const result = state.lastAsk.success ? `got ${state.lastAsk.cardsReceived.length}!` : 'Go Fish!';
    lines.push(`${asker} asked ${target} for ${rankName(state.lastAsk.rank)}: ${result}`);
    if (state.lastAsk.bookFormed) lines.push(`  Book of ${rankName(state.lastAsk.bookRank)} completed!`);
  }

  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentTurn, state.players[state.currentTurn]?.isHuman ?? false);
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
