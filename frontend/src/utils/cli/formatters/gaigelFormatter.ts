import type { GaigelResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];

/** Format a Gaigel game state as terminal text. */
export function formatGaigelState(state: GaigelResponse): string {
  const lines = [formatHeader('Gaigel')];
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(
    `trump: ${state.trumpSuit || 'none'}  stock: ${state.stockRemaining}${state.isEndgame ? ' (endgame)' : ''}`,
  );
  lines.push(
    `scores: team0=${state.teamScores[0]}  team1=${state.teamScores[1]} (first to ${state.config.targetScore})`,
  );
  lines.push(`round points: team0=${state.roundPoints[0]}  team1=${state.roundPoints[1]}`);
  lines.push('');

  for (const player of state.players) {
    lines.push(
      `${formatPlayerName(player.id, player.isHuman)} (team ${player.team}): cards=${player.cardCount} tricks=${player.trickCount}`,
    );
    if (player.isHuman && player.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(player.cards)}  (use p <idx> to play)`);
    }
  }
  lines.push('----------');

  if (state.trumpCard) lines.push(`trump card: ${formatCard(state.trumpCard)}`);
  if (state.currentTrick.length > 0) {
    lines.push(
      `trick: ${state.currentTrick.map((tc) => `${formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false)}=${formatCard(tc.card)}`).join(', ')}`,
    );
  }
  if (state.hint && isRequestedHint(state)) {
    lines.push(
      state.hint.isMarriage
        ? `HINT: declare a marriage with card [${state.hint.cardIndex ?? '-'}] (${state.hint.reason})`
        : `HINT: play card [${state.hint.cardIndex ?? '-'}] (${state.hint.reason})`,
    );
  }
  if (state.message) lines.push(state.message);
  if (state.gameEndFlag)
    lines.push(state.winnerTeam >= 0 ? `Game Over! Team ${state.winnerTeam} wins!` : 'Game Over! Draw.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
