import type { KempsResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Exchange', 'Declare', 'RoundEnd', 'GameEnd'];
const SIGNAL_NAMES = ['Sound', 'Blink'];

/** Format a Kemps game state as terminal text. */
export function formatKempsState(state: KempsResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Kemps'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(
    `Team A: ${state.teamScores[0] ?? 0}  Team B: ${state.teamScores[1] ?? 0}  (target: ${state.targetScore})`,
  );
  lines.push(`your signal: ${SIGNAL_NAMES[state.signalType] ?? state.signalType}`);
  lines.push('----------');

  lines.push(`field: ${formatIndexedCards(state.field)}`);
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    const team = p.team === 0 ? 'A' : 'B';
    const four = p.isHuman && p.hasFourOfAKind ? ' (four of a kind!)' : '';
    lines.push(`${name} [Team ${team}]: ${p.handSize} cards${four}`);
    if (p.hand.length > 0) {
      lines.push(`  ${formatIndexedCards(p.hand)}`);
    }
  });
  lines.push('----------');

  if (state.fourHolderIdx >= 0) {
    lines.push(`four of a kind by: ${formatPlayerName(state.fourHolderIdx, state.fourHolderIdx === 0)}`);
  }

  if (state.phase === 1) {
    if (state.partnerSignaling) {
      lines.push('(your partner is signaling! type "k" to declare Kemps)');
    } else if (state.opponentSignaling) {
      lines.push('(an opponent may be signaling... type "c <seat>" for Counter-Kemps)');
    }
    lines.push('(declare window: k=kemps, c <seat>=counter, p=decline)');
  } else if (state.phase === 0 && state.isHumanTurn) {
    lines.push('(your turn — swap a card with "s <h> <f>" or pass with "p")');
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerTeam >= 0) {
    lines.push(`Game Over! Winner: Team ${state.winnerTeam === 0 ? 'A' : 'B'}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
