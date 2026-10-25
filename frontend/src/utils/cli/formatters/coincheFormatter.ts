import type { CoincheResponse } from '../../../types/card';
import { suitSymbolAt } from '../../cardAlt';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Double', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];

/** Format a Coinche response as readable terminal text. */
export function formatCoincheState(state: CoincheResponse): string {
  const lines = [
    formatHeader('Coinche'),
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
    `scores: team0=${state.teamScores[0]}  team1=${state.teamScores[1]}`,
    `contract: ${state.contractPoints ?? 'none'}  trump: ${state.trumpSuit ?? 'none'}`,
    '',
  ];
  lines.push('Commands: p <idx> plays the card at that zero-based hand index.');
  for (const p of state.players) {
    lines.push(`${formatPlayerName(p.id, p.isHuman)} (team ${p.team}): cards=${p.cards.length} tricks=${p.trickCount}`);
    if (p.isHuman && p.cards.length) lines.push(`  ${p.cards.map((card, i) => `[${i}]${formatCard(card)}`).join(' ')}`);
  }
  lines.push('----------');
  if (state.currentTrick.length)
    lines.push(
      `trick: ${state.currentTrick.map((entry) => `${formatPlayerName(entry.playerIdx, state.players[entry.playerIdx]?.isHuman ?? false)}=${formatCard(entry.card)}`).join(', ')}`,
    );
  if (state.hint && isRequestedHint(state)) {
    if (state.hint.bid !== undefined) {
      const suit = state.hint.suit === undefined ? '' : ` ${suitSymbolAt(state.hint.suit)}`;
      lines.push(`HINT: bid ${state.hint.bid}${suit} (${state.hint.reason})`);
    } else if (state.hint.cardIndex !== undefined) {
      lines.push(`HINT: card index ${state.hint.cardIndex} (${state.hint.reason})`);
    }
  }
  if (state.message) lines.push(state.message);
  if (state.gameEndFlag)
    lines.push(state.winnerTeam >= 0 ? `Game Over! Team ${state.winnerTeam} wins!` : 'Game Over! Draw.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
