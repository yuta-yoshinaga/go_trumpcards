import type { ReversisPlayer, ReversisResponse } from '../../../types/card';
import { ReversisPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [ReversisPhase.PLAY]: 'PLAY',
  [ReversisPhase.ROUND_END]: 'ROUND END',
  [ReversisPhase.GAME_END]: 'GAME END',
};

/** Tricks per round (12 cards each on a 48-card pack). */
const TRICKS_PER_ROUND = 12;

/** Which marked cards a seat has been landed with this round. */
function markStr(p: ReversisPlayer): string {
  const marks = [p.tookQuinola && 'J♥', p.tookDiamondAce && 'A♦'].filter(Boolean);
  return marks.length > 0 ? marks.join(',') : 'clean';
}

/** Format a Reversis game state as terminal text. */
export function formatReversisState(state: ReversisResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Reversis'));
  lines.push(
    `round ${state.roundNumber}/${state.config.rounds} | trick ${state.trickNumber + 1}/${TRICKS_PER_ROUND} | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  // プールと失点配分は盤面から読み取れないので常時出す。
  lines.push(`pool: ${state.pool} chips (taken whole by the fewest penalty points)`);
  lines.push('no trump | A=4 K=3 Q=2 J=1 | J♥ and A♦ cost 5 more and 5 chips');
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
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}: ${p.chips} chips | ${p.roundPenalty} penalty [${markStr(p)}] | ${p.cardCount} cards`,
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
