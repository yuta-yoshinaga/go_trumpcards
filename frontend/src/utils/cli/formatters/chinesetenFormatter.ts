import type { ChineseTenCard, ChineseTenPlayer, ChineseTenResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

/** Render one card, annotating the ones that actually score. */
function formatCtCard(c: ChineseTenCard): string {
  return c.points > 0 ? `${formatCard(c)}(${c.points})` : formatCard(c);
}

function formatList(cards: ChineseTenCard[], indexed: boolean): string {
  if (cards.length === 0) return '-';
  return cards.map((c, i) => (indexed ? `${i}:${formatCtCard(c)}` : formatCtCard(c))).join(' ');
}

/** Render one seat. A hidden seat arrives with a count and no hand cards. */
function formatSeat(p: ChineseTenPlayer): string {
  const who = p.isHuman ? 'you' : 'cpu';
  const hand = p.hidden ? `${p.cardCount} cards` : formatList(p.cards, true);
  // Captures print for BOTH seats: they are public, and reading them is the game.
  return `${who}: ${p.score} pt\n  captured: ${formatList(p.captured, false)}\n  hand: ${hand}`;
}

/** Format a Chinese Ten game state as terminal text. */
export function formatChineseTenState(state: ChineseTenResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Chinese Ten'));
  lines.push(`stock ${state.stockCount} · tie score ${state.tieScore}`);
  // The two capture rules do not overlap and are the thing players get wrong,
  // so they print every frame rather than once at the top.
  lines.push('capture: A-9 sum to ten, 10-K by rank / only red cards score');
  lines.push(`layout: ${formatList(state.layout, true)}`);

  for (const p of state.players) {
    lines.push(formatSeat(p));
  }

  if (state.pendingCard) {
    lines.push(`played: ${formatCtCard(state.pendingCard)} -- choose a layout card`);
  }

  if (state.gameEndFlag) {
    lines.push(state.winnerIdx === 0 ? 'you win' : state.winnerIdx < 0 ? 'draw' : 'you lose');
  }

  return lines.join('\n');
}
