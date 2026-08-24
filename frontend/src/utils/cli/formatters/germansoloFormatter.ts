import type { GermanSoloResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

// **Indexed by the phase value, so AceCall has to be in the list.** Leaving it
// out shifts every later phase down one and the terminal calls TrickEnd "Play".
const PHASE_NAMES = ['Bid', 'AceCall', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const BID_NAMES = ['pass', 'Mussfrage', 'Frage', 'Solo', 'Tout'];
const SUIT_NAMES = ['-', 'spade', 'club', 'heart', 'diamond'];
const OUTCOME_NAMES = ['-', 'made', 'failed'];

/**
 * One line describing the ace call.
 *
 * **The called ace is public, its holder is not.** The shout is heard at the
 * table, but who answers it stays hidden until that ace is played.
 */
function aceCallLine(state: GermanSoloResponse): string {
  if (state.playsAlone) return 'ace call: playing alone';
  if (state.calledAceSuit < 1) return 'ace call: not yet named';
  const suit = SUIT_NAMES[state.calledAceSuit] ?? '-';
  if (state.partnerIdx < 0) return `called ace: A of ${suit} (holder hidden)`;
  return `called ace: A of ${suit} -> partner is P${state.partnerIdx}`;
}

/** Role label for a seat: declarer, the revealed partner, or a defender. */
function germanSoloRole(state: GermanSoloResponse, seat: number): string {
  if (seat === state.declarerIdx) return 'Declarer';
  if (seat === state.partnerIdx) return 'Partner';
  return 'Defender';
}

/** Format a German Solo game state as terminal text. */
export function formatGermanSoloState(state: GermanSoloResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('German Solo'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  const trumpName = state.trumpSuit >= 1 ? (SUIT_NAMES[state.trumpSuit] ?? '-') : '-';
  lines.push(`bid: ${BID_NAMES[state.winningBid] ?? state.winningBid}  trump: ${trumpName}`);
  if (state.declarerIdx >= 0) {
    lines.push(
      `need ${state.requiredTricks} tricks — declarers ${state.declarerTricks} / defenders ${state.defenderTricks}`,
    );
  }
  lines.push(aceCallLine(state));
  lines.push(`scores: ${state.playerScores.map((s, i) => `P${i}=${s}`).join('  ')}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = germanSoloRole(state, p.id);
    lines.push(`${name} (${role}): cards=${p.cardCount} tricks=${p.trickCount} score=${p.score}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trickParts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${trickParts.join(', ')}`);
  }

  if ((state.phase === 4 || state.phase === 5) && state.outcome > 0) {
    lines.push(`round result: ${OUTCOME_NAMES[state.outcome] ?? state.outcome}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerPlayer >= 0) {
    lines.push(`Game Over! Winner: Player ${state.winnerPlayer}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
