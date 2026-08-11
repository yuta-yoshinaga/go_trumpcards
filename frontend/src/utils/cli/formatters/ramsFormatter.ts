import type { RamsPlayer, RamsResponse } from '../../../types/card';
import { RamsPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [RamsPhase.DECIDE]: 'DECIDE',
  [RamsPhase.PLAY]: 'PLAY',
  [RamsPhase.ROUND_END]: 'ROUND END',
  [RamsPhase.GAME_END]: 'GAME END',
};

/** Tricks per round (five cards each). */
const TRICKS_PER_ROUND = 5;

/** Whether a seat is in, out, or has yet to choose. */
function statusStr(p: RamsPlayer): string {
  if (!p.decided) return '-';
  return p.inRound ? 'in' : 'out';
}

/** Format a Rams game state as terminal text. */
export function formatRamsState(state: RamsResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Rams'));
  lines.push(
    `round ${state.roundNumber}/${state.config.rounds} | trick ${state.trickNumber + 1}/${TRICKS_PER_ROUND} | ${
      state.config.playerCnt
    } players | ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  // ポット・切り札・リスクの3点が参加判断の材料。盤面からは読めない。
  lines.push(`pot: ${state.pot} chips (shared out by tricks taken)`);
  if (state.upCard) lines.push(`trump: ${formatCard(state.upCard)}`);
  lines.push('play and take no trick at all and you pay 5 more into the pot');
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
      `${marker}${formatPlayerName(p.id, p.isHuman)}: ${p.chips} chips [${statusStr(p)}] ${p.roundTricks} taken | ${p.cardCount} cards`,
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
