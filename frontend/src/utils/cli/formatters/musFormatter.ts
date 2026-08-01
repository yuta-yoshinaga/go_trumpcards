import type { MusResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['Mus', 'Discard', 'Grande', 'Chica', 'Pares', 'Juego', 'Showdown', 'RoundEnd', 'GameEnd'];
const ROUND_NAMES = ['Grande', 'Chica', 'Pares', 'Juego'];

/** Format a Mus game state as terminal text. */
export function formatMusState(state: MusResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Mus'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(`amarrakos: team0=${state.amarrakos[0] ?? 0}  team1=${state.amarrakos[1] ?? 0}`);
  if (state.pendingStake !== 0) {
    lines.push(`pending bet: ${state.pendingStake === -1 ? 'órdago' : state.pendingStake}`);
  }
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: cards=${p.cardCount} teamScore=${p.teamScore}`);
    if (p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  const resultParts = state.results
    .filter((r) => r.team >= 0)
    .map((r) => `${ROUND_NAMES[r.kind] ?? r.kind}: team${r.team} +${r.stake}`);
  if (resultParts.length > 0) {
    lines.push(`results: ${resultParts.join(', ')}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const ind = state.hint.indices ?? [];
    const indStr = ind.length > 0 ? ` indices [${ind.join(', ')}]` : '';
    lines.push(`HINT: action=${state.hint.action}${indStr} (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerTeam >= 0) {
    lines.push(`Game Over! Winner: team${state.winnerTeam}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
