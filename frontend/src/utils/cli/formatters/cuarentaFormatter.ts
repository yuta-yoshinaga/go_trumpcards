import type { CuarentaAction, CuarentaResponse } from '../../../types/card';
import {
  formatCard,
  formatCardList,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
} from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'Play',
  1: 'RoundEnd',
  2: 'GameEnd',
};

/** Team label for a team index (0 = A, 1 = B). */
function teamName(team: number): string {
  return team === 0 ? 'A' : 'B';
}

/** Render a single play action as a one-line summary with bonus tags. */
function formatAction(a: CuarentaAction): string {
  const name = formatPlayerName(a.playerIdx, a.playerIdx === 0);
  const played = a.playedCard ? formatCard(a.playedCard) : '-';
  const tags: string[] = [];
  if (a.isCaida) tags.push('Caída!+2');
  if (a.rondaBonus > 0) tags.push(`Ronda!+${a.rondaBonus}`);
  if (a.isLimpia) tags.push('Limpia!+1');
  const suffix = tags.length > 0 ? ` (${tags.join(' ')})` : '';
  if (a.capturedCards.length > 0) {
    return `${name}: played ${played} -> captured ${formatCardList(a.capturedCards)}${suffix}`;
  }
  return `${name}: laid ${played}`;
}

/** Format a Cuarenta game state as terminal text. */
export function formatCuarentaState(state: CuarentaResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Cuarenta'));
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? state.phase}  stock: ${state.remainingDeck}`);

  state.teamScores.forEach((score, team) => {
    lines.push(`Team ${teamName(team)}: ${score} / ${state.config.targetScore}`);
  });
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    const turn = i === state.currentTurn && !state.gameEndFlag ? ' <- turn' : '';
    lines.push(`${name} (Team ${teamName(p.team)}): hand ${p.cardCount} / captured ${p.capturedCount}${turn}`);
  });
  lines.push('----------');

  lines.push(`table: ${state.tableCards.length > 0 ? formatCardList(state.tableCards) : '-'}`);

  if (state.humanAction) lines.push(formatAction(state.humanAction));
  for (const a of state.cpuActions) lines.push(formatAction(a));

  const human = state.players.find((p) => p.isHuman);
  if (human && human.cards.length > 0) {
    lines.push(`your hand: ${formatIndexedCards(human.cards)}`);
  }

  if (state.phase === 0 && state.currentTurn === 0 && !state.gameEndFlag) {
    lines.push('(your turn — play a card with "p <hand#>")');
  }

  if (state.message) lines.push(state.message);

  if (state.gameEndFlag) {
    lines.push('Game Over!');
    state.teamScores.forEach((score, team) => {
      lines.push(`  Team ${teamName(team)}: ${score} pts`);
    });
    if (state.roundWinners.length > 0) {
      const names = state.roundWinners.map((t) => `Team ${teamName(t)}`).join(', ');
      lines.push(`Winner: ${names}`);
    }
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
