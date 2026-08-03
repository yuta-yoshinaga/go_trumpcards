import type { Card, NainJaunePlayer, NainJauneResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

function formatList(cards: Card[], indexed: boolean): string {
  if (cards.length === 0) return '-';
  return cards.map((c, i) => (indexed ? `${i}:${formatCard(c)}` : formatCard(c))).join(' ');
}

/** Render one seat. A hidden seat arrives with a count and no hand cards. */
function formatSeat(p: NainJaunePlayer): string {
  const who = p.isHuman ? 'you' : `cpu${p.id.toString()}`;
  const hand = p.hidden ? `${p.cardCount.toString()} cards` : formatList(p.cards, true);
  // 支払いは枚数ではなく点数なので、点も出さないと負債額が読めない。
  return `${who}: ${p.chips.toString()} chips, ${p.cardCount.toString()} card(s) worth ${p.points.toString()}\n  hand: ${hand}`;
}

/** Format a Le Nain Jaune game state as terminal text. */
export function formatNainJauneState(state: NainJauneResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Le Nain Jaune'));
  lines.push(
    `deal ${(state.dealNo + 1).toString()}/${state.targetDeals.toString()} · talon ${state.talonCount.toString()}`,
  );
  // 「スート無関係」と「精算は点数」が最も誤解されやすい。
  lines.push('the run climbs by rank and IGNORES SUIT / a king ends it / you pay in POINTS, not cards');
  // 各区画は「取る札」つきで出す。スートまで一致していないと取れない。
  lines.push(
    `board: ${state.boxes
      .map((b) => `${b.name}(${b.card ? formatCard(b.card) : '?'}):${b.chips.toString()}`)
      .join(' ')}`,
  );

  for (const a of state.awards) {
    lines.push(`seat ${a.player.toString()} takes ${a.box} (${a.chips.toString()})`);
  }

  if (state.playedPile.length > 0) {
    lines.push(`played: ${formatList(state.playedPile, false)}`);
  }
  if (state.runRank === 0) {
    lines.push('the run is stopped; lead any card');
  } else {
    lines.push(`next up: a ${(state.runRank + 1).toString()} of any suit`);
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
