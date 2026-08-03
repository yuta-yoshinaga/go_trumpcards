import { type TarocchiniResponse, tarocchiniTeamOf } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Scarto', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];

/** Format a Tarocchini game state as terminal text. */
export function formatTarocchiniState(state: TarocchiniResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Tarocchini'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`scores: team0=${state.teamScores[0]}  team1=${state.teamScores[1]}`);
  // Without this the four papi look like ordinary trumps and the whole point of
  // holding one back is invisible.
  lines.push('papi: the four rank equal — if more than one lands in a trick, the LATER-played wins');
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const dealer = p.isDealer ? ' [dealer]' : '';
    lines.push(`${name} (team ${p.team}): cards=${p.cardCount} tricks=${p.trickCount}${dealer}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.scartoCount > 0) {
    lines.push(`scarto: ${state.scartoCount} cards buried by ${formatPlayerName(state.dealerIdx, false)}`);
  }

  if (state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  if (state.phase === 3 || state.phase === 4) {
    const tricks = state.roundTricks.map((v, i) => `P${i}(T${tarocchiniTeamOf(i)})=${v}`).join(' ');
    lines.push(`round result: tricks ${tricks}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    // A draw leaves winnerTeam at -1; "Team -1 wins" would be a lie.
    lines.push(state.winnerTeam >= 0 ? `Game Over! Team ${state.winnerTeam} wins!` : 'Game Over! Draw.');
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
