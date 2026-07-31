import type { Card, DesmochePlayer, DesmocheResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

/** Meld kinds, indexed by the wire value. */
const MELD_KINDS = ['set', 'run'];

function formatList(cards: Card[], indexed: boolean): string {
  if (cards.length === 0) return '-';
  return cards.map((c, i) => (indexed ? `${i}:${formatCard(c)}` : formatCard(c))).join(' ');
}

/** Render one seat. A hidden seat arrives with a count and no hand cards. */
function formatSeat(p: DesmochePlayer, goOutSize: number): string {
  const who = p.isHuman ? 'you' : `cpu${p.id.toString()}`;
  const hand = p.hidden ? `${p.cardCount.toString()} cards` : formatList(p.cards, true);
  return `${who}: ${p.score.toString()} chips, ${p.meldedCount.toString()}/${goOutSize.toString()} down\n  hand: ${hand}`;
}

/** Format a Desmoche game state as terminal text. */
export function formatDesmocheState(state: DesmocheResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Desmoche'));
  lines.push(
    `round ${(state.roundNo + 1).toString()} · stock ${state.stockCount.toString()} · pot ${state.pot.toString()}`,
  );
  // Going out takes ten, not the nine dealt, and poker rankings never enter
  // into it. Those are the two things a player gets wrong, so they print every
  // frame.
  lines.push(`melding exactly ${state.goOutSize.toString()} cards takes the pot / poker hand rankings play no part`);
  lines.push(`discard: ${state.discardTop ? formatCard(state.discardTop) : '-'}`);

  state.melds.forEach((m, i) => {
    lines.push(
      `[${i.toString()}] ${MELD_KINDS[m.kind] ?? '?'} (seat ${m.owner.toString()}): ${formatList(m.cards, false)}`,
    );
  });

  for (const p of state.players) {
    lines.push(formatSeat(p, state.goOutSize));
  }

  if (state.roundWinner >= 0) {
    lines.push(`seat ${state.roundWinner.toString()} melded ten and took the pot`);
  } else if (state.roundExhausted) {
    // Without this the growing pot looks like a bug.
    lines.push(`nobody got down to ten; the pot of ${state.pot.toString()} carries over`);
  }

  if (state.gameEndFlag) {
    lines.push(state.winnerIdx === 0 ? 'you finish ahead' : 'you finish behind');
  }

  return lines.join('\n');
}
