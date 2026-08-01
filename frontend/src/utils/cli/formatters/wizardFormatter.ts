import type { WizardResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

/** Format a Wizard game state as terminal text. */
export function formatWizardState(state: WizardResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Wizard'));
  lines.push(
    `round: ${state.roundNumber}/${state.totalRounds}  trick: ${state.trickNumber}  hand size: ${state.handSize}`,
  );
  if (state.trumpCard) lines.push(`trump: ${formatCard(state.trumpCard)}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: total=${p.cumulativeScore} round=${p.roundScore} bid=${p.bid} tricks=${p.trickCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  if (state.phase === 0) {
    lines.push('Bidding phase');
  }

  if (state.hint && isRequestedHint(state)) {
    if (state.hint.bid !== undefined) lines.push(`HINT: bid ${state.hint.bid} (${state.hint.reason})`);
    if (state.hint.cardIndex !== undefined) lines.push(`HINT: play [${state.hint.cardIndex}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winner}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
