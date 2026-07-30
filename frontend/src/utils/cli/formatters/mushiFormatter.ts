import type { MushiCard, MushiResponse } from '../../../types/card';
import { formatHeader } from '../formatterBase';

/** Categories, indexed by the server's `category` field. */
const CATEGORY = ['chaff', 'ribbon', 'animal', 'bright'];

/**
 * Render one hanafuda card as `month-index(category)`. Hanafuda has no PNG art
 * and no suit, so the month and the category are what a reader needs. A `*`
 * marks the lightning card.
 */
function formatMushiCard(c: MushiCard): string {
  return `${c.month}-${c.index}(${CATEGORY[c.category] ?? '?'})${c.isWild ? '*' : ''}`;
}

function formatList(cards: MushiCard[], indexed: boolean): string {
  if (cards.length === 0) return '-';
  return cards.map((c, i) => (indexed ? `${i}:${formatMushiCard(c)}` : formatMushiCard(c))).join(' ');
}

/** Format a Mushi game state as terminal text. */
export function formatMushiState(state: MushiResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Mushi'));
  lines.push(`round ${state.roundNumber}/${state.targetRounds} · stock ${state.stockCount}`);
  lines.push(`field: ${formatList(state.field, true)}`);

  for (const p of state.players) {
    const who = p.isHuman ? 'you' : 'cpu';
    lines.push(`${who}: ${p.score} total (${p.capturedPoints} pt captured / ${p.cardCount} in hand)`);
    // Captured cards print for BOTH seats: they are public, and reading them
    // is how the game is played.
    lines.push(`  captured: ${formatList(p.captured, false)}`);
    if (!p.hidden) {
      lines.push(`  hand: ${formatList(p.cards, true)}`);
    }
  }

  if (state.pendingCard) {
    lines.push(`played: ${formatMushiCard(state.pendingCard)} -- choose a field card`);
  }

  if (state.gameEndFlag) {
    lines.push(state.winnerIdx === 0 ? 'you win' : state.winnerIdx < 0 ? 'draw' : 'you lose');
  }

  return lines.join('\n');
}
