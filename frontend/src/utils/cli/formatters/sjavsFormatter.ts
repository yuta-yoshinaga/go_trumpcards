import type { Card, SjavsPlayer, SjavsResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

const SUIT_NAMES = ['spade', 'club', 'heart', 'diamond'];

function formatList(cards: Card[], indexed: boolean): string {
  if (cards.length === 0) return '-';
  return cards.map((c, i) => (indexed ? `${i}:${formatCard(c)}` : formatCard(c))).join(' ');
}

/** Render one seat. A hidden seat arrives with a count and no hand cards. */
function formatSeat(p: SjavsPlayer): string {
  const who = p.isHuman ? 'you' : `cpu${p.id.toString()}`;
  const hand = p.hidden ? `${p.cardCount.toString()} cards` : formatList(p.cards, true);
  const bid = p.bid > 0 ? ` bid ${p.bid.toString()}` : '';
  return `${who} [team ${p.team.toString()}]${bid}: ${hand}`;
}

/** Format a Sjavs game state as terminal text. */
export function formatSjavsState(state: SjavsResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Sjavs'));
  // Trump is only known after bidding, so naming a suit early would invent one.
  const trump = state.trumpSuit >= 0 ? (SUIT_NAMES[state.trumpSuit] ?? '?') : 'undecided';
  lines.push(
    `trump ${trump} · to go: us ${state.remaining[0]?.toString() ?? '?'} / them ${state.remaining[1]?.toString() ?? '?'}`,
  );
  // The permanent trumps are the shape of the game, so they print every frame:
  // believing only the trump suit is trump is the standard mistake.
  lines.push('permanent trumps: CQ > SQ > CJ > SJ > HJ > DJ (highest whatever the trump suit)');
  if (state.trumpCount > 0) {
    lines.push(`trumps in this suit: ${state.trumpCount.toString()}`);
  }
  lines.push(
    `this hand: us ${state.teamPoints[0]?.toString() ?? '0'} / them ${state.teamPoints[1]?.toString() ?? '0'} (120 in all)`,
  );

  if (state.trick.length > 0) {
    lines.push(
      `trick: ${formatList(
        state.trick.map((tc) => tc.card),
        false,
      )}`,
    );
  }

  for (const p of state.players) {
    lines.push(formatSeat(p));
  }

  if (state.handResult) {
    const r = state.handResult;
    lines.push(
      r.scoringTeam < 0
        ? '60-60: nobody scores, the next game is worth two more'
        : `team ${r.scoringTeam.toString()} takes ${r.amount.toString()} off${r.vol ? ' (all tricks)' : ''}`,
    );
  }

  if (state.gameEndFlag) {
    const won = state.winnerTeam === 0;
    lines.push(
      `${won ? 'you won the rubber' : 'you lost the rubber'}${state.doubleVictory ? ' (double victory)' : ''}`,
    );
  }

  return lines.join('\n');
}
