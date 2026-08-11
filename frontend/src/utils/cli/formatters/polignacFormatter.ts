import type { PolignacResponse } from '../../../types/card';
import { PolignacPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [PolignacPhase.DECLARE]: 'DECLARE',
  [PolignacPhase.PLAY]: 'PLAY',
  [PolignacPhase.ROUND_END]: 'ROUND END',
  [PolignacPhase.GAME_END]: 'GAME END',
};

/** Tricks per round — capot means taking all of them. */
const TRICKS_PER_ROUND = 8;

/** Format a Polignac game state as terminal text. */
export function formatPolignacState(state: PolignacResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Polignac'));
  lines.push(
    `round ${state.roundNumber}/${state.config.rounds} | trick ${state.trickNumber + 1}/${TRICKS_PER_ROUND} | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  // 盤面からは読み取れない規則なので常時出す。
  lines.push('no trump | each jack costs 1, the J of spades costs 2');

  if (state.capotIdx >= 0) {
    const who = formatPlayerName(state.capotIdx, state.players[state.capotIdx]?.isHuman ?? false);
    lines.push(`! ${who} declared capot (${state.capotTricks}/${TRICKS_PER_ROUND}) — one trick breaks it`);
  }

  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trick = state.currentTrick
      .map((tc) => `${formatPlayerName(tc.playerIdx, false)}:${formatCard(tc.card)}`)
      .join('  ');
    lines.push(`trick: ${trick}`);
    lines.push('----------');
  }

  state.players.forEach((p) => {
    const marker = p.id === state.currentPlayerIdx && !state.gameEndFlag ? '>' : ' ';
    const capot = p.declaredCapot ? ' [capot]' : '';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}${capot}: ${p.score} pts (round ${p.roundPenalty}) ${p.cardCount} cards`,
    );
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
    lines.push(
      state.winnerIdx >= 0
        ? `game over — winner: ${formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false)}`
        : 'game over — tie',
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
