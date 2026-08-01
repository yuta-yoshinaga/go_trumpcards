import type { OmbreResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const BID_NAMES = ['pass', 'entrar', 'solo'];
const SUIT_NAMES = ['-', 'spade', 'club', 'heart', 'diamond'];
const OUTCOME_NAMES = ['-', 'Sacar', 'Puesta', 'Codille'];

/** Format an Ombre (Hombre) game state as terminal text. */
export function formatOmbreState(state: OmbreResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Ombre'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  const trumpName = state.trumpSuit >= 1 ? (SUIT_NAMES[state.trumpSuit] ?? '-') : '-';
  lines.push(`bid: ${BID_NAMES[state.winningBid] ?? state.winningBid}  trump: ${trumpName}`);
  lines.push(`scores: ${state.playerScores.map((s, i) => `P${i}=${s}`).join('  ')}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isOmbre ? 'Ombre' : 'Coalition';
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

  if ((state.phase === 3 || state.phase === 4) && state.outcome > 0) {
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
