import type { SlobberhannesPlayer, SlobberhannesResponse } from '../../../types/card';
import { SlobberhannesPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [SlobberhannesPhase.PLAY]: 'PLAY',
  [SlobberhannesPhase.ROUND_END]: 'ROUND END',
  [SlobberhannesPhase.GAME_END]: 'GAME END',
};

/** Tricks per round — the last one carries a penalty, so the count is load-bearing. */
const TRICKS_PER_ROUND = 8;

/** Compact marks for the three penalties a player may already have taken. */
function penaltyMarks(p: SlobberhannesPlayer): string {
  const marks = [p.tookFirstTrick && '1st', p.tookLastTrick && 'last', p.tookQueen && 'Q♣'].filter(Boolean);
  return marks.length > 0 ? marks.join(',') : 'clean';
}

/** Format a Slobberhannes game state as terminal text. */
export function formatSlobberhannesState(state: SlobberhannesResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Slobberhannes'));
  lines.push(
    `round ${state.roundNumber}/${state.config.rounds} | trick ${state.trickNumber + 1}/${TRICKS_PER_ROUND} | ${
      PHASE_NAMES[state.phase] ?? state.phase
    }`,
  );
  lines.push('no trump: the highest card of the led suit takes the trick');

  // 位置そのものが罰の対象。中身に関係なく警告する。
  if (state.trickNumber === 0) lines.push('! FIRST trick — taking it costs 1 point');
  else if (state.trickNumber === TRICKS_PER_ROUND - 1) lines.push('! LAST trick — taking it costs 1 point');

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
      `${marker}${formatPlayerName(p.id, p.isHuman)}: ${p.score} pts [${penaltyMarks(p)}] ${p.cardCount} cards`,
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
