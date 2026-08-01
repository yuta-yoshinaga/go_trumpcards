import type { CinchResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Bid', 'NameTrump', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const SUIT_SYMBOLS = ['-', '♠', '♣', '♥', '♦'];

/** Format a Cinch (Double Pedro) game state as terminal text. */
export function formatCinchState(state: CinchResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Cinch'));
  lines.push(
    `deal: ${state.roundNumber}  trick: ${state.trickNumber}/${state.totalTricks}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`high bid: ${state.currentBid}  trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '-'}`);
  lines.push(`scores: ${state.players.map((p) => `P${p.id}=${p.totalScore}`).join('  ')}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const badge = p.id === state.bidWinnerIdx ? ' (Bidder)' : '';
    lines.push(`${name}${badge}: cards=${p.cardCount} tricks=${p.trickCount} bid=${p.bid} score=${p.totalScore}`);
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

  if ((state.phase === 4 || state.phase === 5) && state.lastDealDetail) {
    const d = state.lastDealDetail;
    if (d.setBack) lines.push('deal result: bidder SET BACK');
    const gained = state.players.map((p) => `P${p.id}=${d.gained[p.id] ?? 0}`).join(' ');
    lines.push(`deal result: gained ${gained}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerIdx >= 0) {
    lines.push(`Game Over! Winner: Player ${state.winnerIdx}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
