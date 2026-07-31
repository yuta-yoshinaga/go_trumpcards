import type { Card, TrexPlayer, TrexResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

const SUIT_NAMES = ['joker', 'spade', 'club', 'heart', 'diamond'];

/** Contract names, indexed by the wire value. Index 5 is "not chosen". */
const CONTRACT_NAMES = [
  'King of Hearts (-75)',
  'Diamonds (-10 each)',
  'Queens (-25 each)',
  'Tricks (-15 each)',
  'Dominoes (+200/+150/+100/+50)',
  'not chosen',
];

function formatList(cards: Card[], indexed: boolean): string {
  if (cards.length === 0) return '-';
  return cards.map((c, i) => (indexed ? `${i}:${formatCard(c)}` : formatCard(c))).join(' ');
}

/** Render one seat. A hidden seat arrives with a count and no hand cards. */
function formatSeat(p: TrexPlayer, kingIdx: number): string {
  const who = p.isHuman ? 'you' : `cpu${p.id.toString()}`;
  const role = p.id === kingIdx ? ' (king)' : '';
  const hand = p.hidden ? `${p.cardCount.toString()} cards` : formatList(p.cards, true);
  return `${who}${role}: ${p.score.toString()} total (deal ${p.dealScore.toString()})\n  hand: ${hand}`;
}

/** Format a Trex game state as terminal text. */
export function formatTrexState(state: TrexResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Trex'));
  lines.push(
    `deal ${state.dealNo.toString()}/${state.totalDeals.toString()} · king seat ${state.kingIdx.toString()} · ${CONTRACT_NAMES[state.contract] ?? '?'}`,
  );
  // Two things a player gets wrong: a contract is played once per kingdom, and
  // the dominoes build from the JACK rather than the seven.
  lines.push('each contract once per kingdom / the dominoes build out from the JACK');

  if (state.availableContracts.length > 0) {
    lines.push(
      `left to choose: ${state.availableContracts.map((c) => `${c.toString()}=${CONTRACT_NAMES[c] ?? '?'}`).join(' / ')}`,
    );
  }

  if (state.isTrix) {
    for (const run of state.runs) {
      const span = run.started ? `${run.low.toString()}-${run.high.toString()}` : 'not started';
      lines.push(`${SUIT_NAMES[run.suit] ?? '?'}: ${span}`);
    }
  } else if (state.trick.length > 0) {
    lines.push(
      `trick: ${formatList(
        state.trick.map((tc) => tc.card),
        false,
      )}`,
    );
  }

  for (const p of state.players) {
    lines.push(formatSeat(p, state.kingIdx));
  }

  if (state.canPass) {
    lines.push('no legal play -- use s to pass');
  }

  if (state.gameEndFlag) {
    lines.push(state.winnerIdx === 0 ? 'you win' : 'you lose');
  }

  return lines.join('\n');
}
