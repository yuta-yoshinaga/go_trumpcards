import type { CucumberResponse } from '../../../types/card';
import { CucumberPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [CucumberPhase.PLAY]: 'PLAY',
  [CucumberPhase.ROUND_END]: 'ROUND END',
  [CucumberPhase.GAME_END]: 'GAME END',
};

/** Format a Cucumber game state as terminal text. */
export function formatCucumberState(state: CucumberResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Cucumber'));
  lines.push(
    `round ${state.roundNumber} | trick ${state.trickNumber + 1}/${state.totalTricks} | ends at ${state.config.targetScore} | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  // **スート無関係・失点は最終トリックだけ、が規則そのもの。**
  lines.push('suits are irrelevant — beat the highest card or dump your lowest; only the last trick scores');

  // **超えるべきランクは盤面から数えさせない。**
  lines.push(state.highestInTrick > 0 ? `rank to beat: ${state.highestInTrick}` : 'you lead — any card is legal');

  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trick = state.currentTrick
      .map((tc) => `${formatPlayerName(tc.playerIdx, false)}:${formatCard(tc.card)}`)
      .join('  ');
    lines.push(`trick: ${trick}`);
    lines.push('----------');
  }

  state.players.forEach((p) => {
    const marker = p.id === state.currentPlayerIdx && state.phase === CucumberPhase.PLAY ? '>' : ' ';
    const role = p.id === state.lastTrickWinnerIdx && state.lastPenalty > 0 ? `[last trick +${state.lastPenalty}]` : '';
    lines.push(`${marker}${formatPlayerName(p.id, p.isHuman)}${role}: ${p.cardCount} cards, ${p.penalty} penalty`);
  });

  const human = state.players.find((p) => p.isHuman);
  if (human) {
    lines.push('----------');
    const hand = human.cards
      .map((c, i) => `[${i}]${formatCard(c)}${state.validPlays.includes(i) ? '*' : ''}`)
      .join('  ');
    lines.push(`your hand: ${hand || '(empty)'}`);
  }

  if (state.gameEndFlag) {
    lines.push('----------');
    const winner = state.players[state.winnerIdx];
    lines.push(
      `game over — ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} finished with the fewest (${winner?.penalty ?? 0})`,
    );
  } else if (state.phase === CucumberPhase.ROUND_END && state.lastTrickWinnerIdx >= 0) {
    lines.push(
      `${formatPlayerName(state.lastTrickWinnerIdx, state.lastTrickWinnerIdx === 0)} took the last trick — ${state.lastPenalty} penalty`,
    );
    lines.push('next — deal the next round');
  } else if (state.forced) {
    // **合法手が1つ = 更新できない、ではない。** サーバの forced をそのまま使う。
    lines.push('you cannot beat it — your lowest card is the only legal play');
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
