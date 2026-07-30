import type { Card, SkitgubbePlayer, SkitgubbeResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

const SUIT_NAMES = ['spade', 'club', 'heart', 'diamond'];

function formatList(cards: Card[], indexed: boolean): string {
  if (cards.length === 0) return '-';
  return cards.map((c, i) => (indexed ? `${i}:${formatCard(c)}` : formatCard(c))).join(' ');
}

/** Render one seat. A hidden seat arrives with a count and no hand cards. */
function formatSeat(p: SkitgubbePlayer): string {
  const who = p.isHuman ? 'you' : `cpu${p.id.toString()}`;
  const hand = p.hidden ? `${p.cardCount.toString()} cards` : formatList(p.cards, true);
  return `${who}: ${p.cardCount.toString()} in hand, ${p.collectedCount.toString()} collected\n  hand: ${hand}`;
}

/** Format a Skitgubbe game state as terminal text. */
export function formatSkitgubbeState(state: SkitgubbeResponse): string {
  const lines: string[] = [];
  const collecting = state.phase === 0;

  lines.push(formatHeader('Skitgubbe'));
  // Trump is fixed by the LAST card drawn, so it is genuinely unknown early on
  // -- printing a suit there would invent one.
  const trump = state.trumpSuit >= 0 ? (SUIT_NAMES[state.trumpSuit] ?? '?') : 'undecided';
  lines.push(`${collecting ? 'phase 1 (collect)' : 'phase 2 (shed)'} · stock ${state.stockCount} · trump ${trump}`);
  // The two phases are different games, so their rules print every frame.
  lines.push('p1: two-player duel, suit irrelevant / p2: beat the pile, or pick it up');

  if (collecting) {
    lines.push(`duel: ${formatList(state.duel, false)} (leader ${state.duelLeader})`);
  } else {
    lines.push(`pile: ${formatList(state.pile, false)}`);
  }

  for (const p of state.players) {
    lines.push(formatSeat(p));
  }

  if (state.canPickUp) {
    lines.push('nothing beats the pile -- use u to pick it up');
  }

  if (state.gameEndFlag) {
    lines.push(state.loserIdx === 0 ? 'you are the skitgubbe' : 'you got rid of your cards');
  }

  return lines.join('\n');
}
