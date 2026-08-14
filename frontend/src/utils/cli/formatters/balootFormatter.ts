import type { BalootResponse } from '../../../types/card';
import { BalootMode, BalootPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [BalootPhase.DECLARE]: 'DECLARE',
  [BalootPhase.PLAY]: 'PLAY',
  [BalootPhase.ROUND_END]: 'ROUND END',
  [BalootPhase.GAME_END]: 'GAME END',
};

/** Tricks per round (eight cards each). */
const TRICKS_PER_ROUND = 8;

/** trumpSuit is a 1-based suit code, as elsewhere in this repo. */
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Format a Baloot game state as terminal text. */
export function formatBalootState(state: BalootResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Baloot'));
  lines.push(
    `round ${state.roundNumber} | trick ${state.trickNumber + 1}/${TRICKS_PER_ROUND} | first to ${
      state.config.target
    } | ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`score: yours=${state.scores[0] ?? 0} theirs=${state.scores[1] ?? 0}`);

  // **序列はモードで入れ替わる。** 有効な方だけを出さないと、同じ札の強さが
  // ラウンドごとに違う理由が読めない。
  const declarer =
    state.declarerIdx >= 0
      ? formatPlayerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false)
      : '';
  if (state.mode === BalootMode.SUN) {
    lines.push(`mode: Sun, no trump (declared by ${declarer})`);
    lines.push('order: A=11 > 10 > K=4 > Q=3 > J=2 > 9 > 8 > 7 (120 a round)');
  } else if (state.mode === BalootMode.HOKOM) {
    lines.push(`mode: Hokom, trump ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'} (declared by ${declarer})`);
    lines.push('trump: J=20 > 9=14 > A=11 > 10 > K=4 > Q=3 > 8 > 7; plain suits as in Sun (152 a round)');
  } else {
    lines.push('mode: undeclared — sun or hokom <suit>');
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
    const baloot = p.hasBaloot ? 'Baloot(K+Q)=20' : 'no bonus';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}[T${p.team}]: ${baloot} | ${p.trickCount} tricks | ${p.cardCount} cards`,
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
      state.winnerTeam >= 0
        ? `game over — team ${state.winnerTeam} wins (${state.scores[0] ?? 0} - ${state.scores[1] ?? 0})`
        : `game over — tie (${state.scores[0] ?? 0} - ${state.scores[1] ?? 0})`,
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
