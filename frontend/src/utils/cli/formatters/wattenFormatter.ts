import type { WattenResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Declare', 'Play', 'Respond', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const SUIT_NAMES = ['-', 'spade', 'club', 'heart', 'diamond'];

/** Formats a Schlag rank value as its short card label (1=A, 11=J, 12=Q, 13=K). */
function rankLabel(rank: number): string {
  if (rank <= 0) return '-';
  if (rank === 1) return 'A';
  if (rank === 11) return 'J';
  if (rank === 12) return 'Q';
  if (rank === 13) return 'K';
  return String(rank);
}

/** Format a Watten (ヴァッテン) game state as terminal text. */
export function formatWattenState(state: WattenResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Watten'));
  lines.push(
    `deal: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );

  const critical = state.criticalSuit >= 1 ? (SUIT_NAMES[state.criticalSuit] ?? '-') : '-';
  lines.push(`schlag: ${rankLabel(state.schlagRank)}  critical: ${critical}`);

  const stakeLine =
    state.pendingStake > 0 ? `stake: ${state.stake} (pending ${state.pendingStake})` : `stake: ${state.stake}`;
  lines.push(stakeLine);
  lines.push(`scores: T0=${state.teamScores[0]}  T1=${state.teamScores[1]}`);
  lines.push(`tricks: T0=${state.teamTricks[0]}  T1=${state.teamTricks[1]}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name} (Team ${p.team}): cards=${p.cardCount} tricks=${p.trickCount}`);
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

  if (state.hint && isRequestedHint(state)) {
    const target = state.hint.cardIndex !== undefined ? ` [${state.hint.cardIndex}]` : '';
    lines.push(`HINT: ${state.hint.action}${target} (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerTeam >= 0) {
    lines.push(`Game Over! Winner: Team ${state.winnerTeam}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
