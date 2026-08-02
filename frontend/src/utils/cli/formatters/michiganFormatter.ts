import type { MichiganResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Bet', 'Play', 'Result'];

/** Format a Michigan (Newmarket) game state as terminal text. */
export function formatMichiganState(state: MichiganResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Michigan'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}  ante: ${state.ante}`);

  // Boodles (center betting cards).
  lines.push('boodles:');
  for (const b of state.boodles) {
    const status = b.claimedBy >= 0 ? `claimed by P${b.claimedBy}` : 'open';
    lines.push(`  ${formatCard(b.card)}  chips=${b.chips} [${status}]`);
  }

  // Current sequence.
  if (state.needNewSequence || state.seqSuit === 0) {
    lines.push('sequence: (start a new run)');
  } else {
    lines.push(`sequence: ${state.seqSuitName} up to ${state.seqHighValue}`);
  }
  lines.push(`dead hand: ${state.deadHandCount} cards`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const status = p.isWinner ? 'WENT OUT' : p.isCurrent ? 'to play' : 'waiting';
    lines.push(`${name}: chips=${p.chips} bet=${p.roundBet} hand=${p.cardCount} [${status}]`);
    if (p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.playableIndices.length > 0) {
    lines.push(`playable: ${state.playableIndices.join(', ')}`);
  }

  if (state.hint && isRequestedHint(state)) {
    lines.push(`HINT: play ${state.hint.cardIndex} (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.matchWinnerIdx >= 0) {
    lines.push(
      `Game Over! Winner: ${formatPlayerName(state.matchWinnerIdx, state.players[state.matchWinnerIdx]?.isHuman ?? false)}`,
    );
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
