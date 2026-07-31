import type { Card, LobaPlayer, LobaResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

/** Meld kinds, indexed by the wire value. */
const MELD_KINDS = ['pierna', 'escalera'];

function formatList(cards: Card[], indexed: boolean): string {
  if (cards.length === 0) return '-';
  return cards.map((c, i) => (indexed ? `${i}:${formatCard(c)}` : formatCard(c))).join(' ');
}

/** Render one seat. A hidden seat arrives with a count and no hand cards. */
function formatSeat(p: LobaPlayer): string {
  const who = p.isHuman ? 'you' : `cpu${p.id.toString()}`;
  const hand = p.hidden ? `${p.cardCount.toString()} cards` : formatList(p.cards, true);
  const out = p.eliminated ? ' -- out' : '';
  return `${who}: ${p.score.toString()} penalty${out}\n  hand: ${hand}`;
}

/** Format a Loba game state as terminal text. */
export function formatLobaState(state: LobaResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Loba'));
  lines.push(
    `round ${(state.roundNo + 1).toString()} · stock ${state.stockCount.toString()} · out at ${state.knockOut.toString()}`,
  );
  // The three-different-suits rule and the joker restriction are what a player
  // gets wrong, so they print every frame.
  lines.push(
    'pierna = one rank in three DIFFERENT suits / escalera = a suit run (one joker at most, never in a pierna)',
  );
  lines.push(`discard: ${state.discardTop ? formatCard(state.discardTop) : '-'}`);

  state.melds.forEach((m, i) => {
    lines.push(
      `[${i.toString()}] ${MELD_KINDS[m.kind] ?? '?'} (seat ${m.owner.toString()}): ${formatList(m.cards, false)}`,
    );
  });

  for (const p of state.players) {
    lines.push(formatSeat(p));
  }

  if (state.roundWinner >= 0) {
    lines.push(`seat ${state.roundWinner.toString()} went out${state.roundClean ? ' in one go (-10)' : ''}`);
  }

  if (state.gameEndFlag) {
    lines.push(state.winnerIdx === 0 ? 'you are the last one standing' : 'you were knocked out');
  }

  return lines.join('\n');
}
