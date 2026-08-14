import type { HokmResponse } from '../../../types/card';
import { HokmPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [HokmPhase.TRUMP]: 'TRUMP',
  [HokmPhase.PLAY]: 'PLAY',
  [HokmPhase.HAND_END]: 'HAND END',
  [HokmPhase.GAME_END]: 'GAME END',
};

/** trumpSuit is a 1-based suit code, as elsewhere in this repo. */
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Format a Hokm game state as terminal text. */
export function formatHokmState(state: HokmResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Hokm'));
  lines.push(
    `hand ${state.handNumber} | first to ${state.config.target} hands | ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  // **7 トリック先取が肝。** 13 まで打たないので、進捗はトリック数のほうに出る。
  lines.push(
    `tricks: yours=${state.teamTricks[0] ?? 0} theirs=${state.teamTricks[1] ?? 0} (first to ${state.tricksToWin} takes the hand)`,
  );
  lines.push(`hand points: yours=${state.scores[0] ?? 0} theirs=${state.scores[1] ?? 0}`);
  lines.push(
    state.trumpSuit > 0
      ? `trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'}`
      : 'trump: undeclared (the hakem chooses from their first five)',
  );

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
    const role = p.isHakem ? '[hakem]' : '';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}[T${p.team}]${role}: ${p.trickCount} tricks | ${p.cardCount} cards`,
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

  // **Kot は 2 点。** 何が起きたのか言わないと得点が飛んで見える。
  if (state.phase === HokmPhase.HAND_END && state.lastHandWinner >= 0) {
    lines.push('----------');
    lines.push(
      state.lastHandKot
        ? `hand over — team ${state.lastHandWinner} took every trick: Kot, 2 points`
        : `hand over — team ${state.lastHandWinner} reached ${state.tricksToWin} tricks`,
    );
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
