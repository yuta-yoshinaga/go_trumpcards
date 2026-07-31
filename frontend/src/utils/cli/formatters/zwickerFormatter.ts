import type { ZwickerCard, ZwickerPlayer, ZwickerResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

/**
 * Render one card with its matching values.
 *
 * **Without the values the game is unplayable**: an ace or court card can be
 * used two ways and a joker is 15/20/25, none of which is readable from the
 * rank.
 */
function formatValued(c: ZwickerCard): string {
  return `${formatCard(c)}(${c.values.join('/')})`;
}

function formatList(cards: ZwickerCard[], indexed: boolean): string {
  if (cards.length === 0) return '-';
  return cards.map((c, i) => (indexed ? `${i}:${formatValued(c)}` : formatValued(c))).join(' ');
}

/** Render one seat. A hidden seat arrives with a count and no hand cards. */
function formatSeat(p: ZwickerPlayer): string {
  const who = p.isHuman ? 'you' : `cpu${p.id.toString()}`;
  const hand = p.hidden ? `${p.cardCount.toString()} cards` : formatList(p.cards, true);
  return `${who} (team ${p.team.toString()}): ${p.capturedCount.toString()} taken, ${p.zwicks.toString()} zwick(s)\n  hand: ${hand}`;
}

/** Format a Zwicker game state as terminal text. */
export function formatZwickerState(state: ZwickerResponse): string {
  const lines: string[] = [];
  const us = state.teamScores[0] ?? 0;
  const them = state.teamScores[1] ?? 0;

  lines.push(formatHeader('Zwicker'));
  lines.push(
    `stock ${state.stockCount.toString()} · us ${us.toString()} them ${them.toString()} (first to ${state.targetScore.toString()})`,
  );
  // The two-value cards and what a Zwick actually is are what a player gets
  // wrong, so they print every frame.
  lines.push('A=1/11 J=2/12 Q=3/13 K=4/14 -- you choose / jokers fixed at 15,20,25 / clearing the table is a Zwick');
  lines.push(`table: ${formatList(state.tableCards, true)}`);

  state.builds.forEach((b, i) => {
    const cards = b.cards.map((c) => formatCard(c)).join(' ');
    lines.push(`build[${i.toString()}] worth ${b.value.toString()} (seat ${b.owner.toString()}): ${cards}`);
  });

  for (const p of state.players) {
    lines.push(formatSeat(p));
  }

  const last = state.lastRound;
  if (last) {
    lines.push(`last deal: us ${(last.total[0] ?? 0).toString()}, them ${(last.total[1] ?? 0).toString()}`);
    // 同数だと 3 点が宙に浮く。黙っていると合計が合わないように見える。
    if (last.majorityTeam < 0) {
      lines.push('the card counts were level, so nobody took the three for the majority');
    }
  }

  if (state.gameEndFlag) {
    lines.push(state.winnerTeam === 0 ? 'your team wins' : 'the other team wins');
  }

  return lines.join('\n');
}
