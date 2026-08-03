import { type MinchiateResponse, minchiateTeamOf } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Scarto', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];

/** Format a Minchiate game state as terminal text. */
export function formatMinchiateState(state: MinchiateResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Minchiate'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`scores: team0=${state.teamScores[0]}  team1=${state.teamScores[1]}`);
  // 40 trumps rather than the usual 21 — how many still outrank you is the whole
  // read, and a bare number line does not convey that.
  lines.push('trumps: 40 (zodiac, elements, virtues) — track what is still above you');
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
    const tricks = state.roundTricks.map((v, i) => `P${i}(T${minchiateTeamOf(i)})=${v}`).join(' ');
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
