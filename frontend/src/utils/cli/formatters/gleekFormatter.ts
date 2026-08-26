import type { GleekResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

// **Indexed by the phase value, so Discard has to be in the list.** Leaving it
// out shifts every later phase down one and the terminal calls Play "Discard".
const PHASE_NAMES = ['Bid', 'Discard', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const SUIT_NAMES = ['-', 'spade', 'club', 'heart', 'diamond'];
const RANK_NAMES: Readonly<Record<number, string>> = { 1: 'aces', 11: 'jacks', 12: 'queens', 13: 'kings' };

/** Format a Gleek game state as terminal text. */
export function formatGleekState(state: GleekResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Gleek'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  const trumpName = state.trumpSuit >= 1 ? (SUIT_NAMES[state.trumpSuit] ?? '-') : '-';
  const turnUp = state.turnUp ? formatCard(state.turnUp) : '-';
  lines.push(`trump: ${trumpName}  turn-up: ${turnUp}`);
  lines.push(stockLine(state));
  if (state.ruffWinnerIdx >= 0) {
    const winner = state.players[state.ruffWinnerIdx];
    const suit = winner ? (SUIT_NAMES[winner.ruffSuit] ?? '-') : '-';
    lines.push(`ruff: P${state.ruffWinnerIdx} with ${winner?.ruff ?? 0} in ${suit}`);
  }
  for (const m of state.melds) {
    const kind = m.count >= 4 ? 'mournival' : 'gleek';
    lines.push(`${kind}: P${m.playerIdx} shows ${m.count} ${RANK_NAMES[m.rank] ?? m.rank} for ${m.value} each`);
  }
  lines.push(`scores: ${state.playerScores.map((s, i) => `P${i}=${s}`).join('  ')}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isBuyer ? 'buyer' : 'opponent';
    lines.push(
      `${name} (${role}): cards=${p.cardCount} tricks=${p.trickCount} points=${p.trickPoints} score=${p.score}`,
    );
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

  if (state.phase === 4 || state.phase === 5) {
    lines.push(`deal points: ${state.dealPoints}  par: ${state.par}`);
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

/** One line describing the stock: who bought it and for how much. */
function stockLine(state: GleekResponse): string {
  if (state.buyerIdx < 0) {
    return `stock: unsold (highest ${state.highestBid}, next ${state.nextBidAmount || '-'})`;
  }
  return `stock: bought by P${state.buyerIdx} for ${state.winningBid}`;
}
