import type { Card, PopeJoanPlayer, PopeJoanResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

function formatList(cards: Card[], indexed: boolean): string {
  if (cards.length === 0) return '-';
  return cards.map((c, i) => (indexed ? `${i}:${formatCard(c)}` : formatCard(c))).join(' ');
}

/** Render one seat. A hidden seat arrives with a count and no hand cards. */
function formatSeat(p: PopeJoanPlayer): string {
  const who = p.isHuman ? 'you' : `cpu${p.id.toString()}`;
  const hand = p.hidden ? `${p.cardCount.toString()} cards` : formatList(p.cards, true);
  // 支払い免除の有無は精算を読むのに要る。
  const pope = p.holdsPope ? ' -- holds the Pope' : '';
  return `${who}: ${p.chips.toString()} chips${pope}\n  hand: ${hand}`;
}

/** Format a Pope Joan game state as terminal text. */
export function formatPopeJoanState(state: PopeJoanResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Pope Joan'));
  lines.push(
    `deal ${(state.dealNo + 1).toString()}/${state.targetDeals.toString()} · turn-up ${state.turnUp ? formatCard(state.turnUp) : '-'}`,
  );
  // Paying only on trumps, and the missing eight of diamonds, are what a player
  // gets wrong -- print them every frame.
  lines.push(
    'a compartment pays only on a trump (the Pope, 9D, excepted) / the 8D is out, so a run always dies at the 7',
  );
  lines.push(`board: ${state.compartments.map((c) => `${c.name}:${c.chips.toString()}`).join(' ')}`);

  for (const a of state.awards) {
    const how = a.byTurnUp ? ' from the turn-up' : '';
    lines.push(`seat ${a.player.toString()} takes ${a.compartment} (${a.chips.toString()})${how}`);
  }

  if (state.playedPile.length > 0) {
    lines.push(`played: ${formatList(state.playedPile, false)}`);
  }
  // 止まっているかどうかで出せる札がまるで違う。
  if (state.runSuit < 0) {
    lines.push('the run is stopped; lead your lowest card of any suit');
  }

  for (const p of state.players) {
    lines.push(formatSeat(p));
  }

  if (state.dealWinner >= 0) {
    lines.push(`seat ${state.dealWinner.toString()} went out`);
  }
  if (state.gameEndFlag) {
    lines.push(state.winnerIdx === 0 ? 'you finish with the most chips' : 'you finish behind');
  }

  return lines.join('\n');
}
