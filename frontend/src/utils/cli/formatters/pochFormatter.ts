import type { Card, PochPlayer, PochResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

function formatList(cards: Card[], indexed: boolean): string {
  if (cards.length === 0) return '-';
  return cards.map((c, i) => (indexed ? `${i}:${formatCard(c)}` : formatCard(c))).join(' ');
}

/** Render one seat. A hidden seat arrives with a count and no hand cards. */
function formatSeat(p: PochPlayer): string {
  const who = p.isHuman ? 'you' : `cpu${p.id.toString()}`;
  const hand = p.hidden ? `${p.cardCount.toString()} cards` : formatList(p.cards, true);
  const folded = p.folded ? ' -- folded' : '';
  return `${who}: ${p.chips.toString()} chips, bet ${p.bet.toString()}${folded}\n  hand: ${hand}`;
}

/** Format a Poch game state as terminal text. */
export function formatPochState(state: PochResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Poch'));
  lines.push(
    `deal ${(state.dealNo + 1).toString()}/${state.targetDeals.toString()} · turn-up ${state.turnUp ? formatCard(state.turnUp) : '-'}`,
  );
  // Paying only on the turn-up's suit, and pochen being a comparison rather
  // than a declaration, are what a player gets wrong -- print them every frame.
  lines.push(
    'a pool pays only on a card of the turn-up suit / pochen compares same-rank sets (4 > 3 > 2), no declaration',
  );
  lines.push(`board: ${state.pools.map((p) => `${p.name}:${p.chips.toString()}`).join(' ')}`);

  for (const a of state.stakingAwards) {
    lines.push(`seat ${a.player.toString()} takes ${a.pool} (${a.chips.toString()})`);
  }

  if (state.playedPile.length > 0) {
    lines.push(`played: ${formatList(state.playedPile, false)}`);
  }
  // 並びが止まっているかどうかで、出せる札がまるで違う。
  if (state.stopsSuit < 0) {
    lines.push('the run is stopped; any card may be led');
  }

  for (const p of state.players) {
    lines.push(formatSeat(p));
  }

  if (state.pochenWinner >= 0) {
    lines.push(`seat ${state.pochenWinner.toString()} won the pochen (${state.pochenPot.toString()})`);
  }
  if (state.dealWinner >= 0) {
    lines.push(`seat ${state.dealWinner.toString()} went out`);
  }
  if (state.gameEndFlag) {
    lines.push(state.winnerIdx === 0 ? 'you finish with the most chips' : 'you finish behind');
  }

  return lines.join('\n');
}
