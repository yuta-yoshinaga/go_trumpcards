import type { LingerLongerResponse } from '../../../types/card';
import { LingerLongerPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [LingerLongerPhase.PLAY]: 'PLAY',
  [LingerLongerPhase.GAME_END]: 'GAME END',
};

/** Format a Linger Longer game state as terminal text. */
export function formatLingerLongerState(state: LingerLongerResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Linger Longer'));
  lines.push(`trick ${state.trickNumber + 1} | stock ${state.stockSize} | ${PHASE_NAMES[state.phase] ?? state.phase}`);
  // **取っても得点にならない規則を毎回書く。** 直感と逆なので。
  lines.push('winning a trick only earns you one card from the stock — the last player still holding cards wins');

  // **山札が尽きたら誰も補充できない。** そこから一気に脱落が進む。
  if (state.stockSize === 0 && !state.gameEndFlag) {
    lines.push('the stock is empty — nobody can refill now');
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
    const role = p.eliminatedAt > 0 ? `[out ${p.eliminatedAt}]` : p.id === state.lastDrawIdx ? '[just drew]' : '';
    lines.push(`${marker}${formatPlayerName(p.id, p.isHuman)}${role}: ${p.cardCount} cards, won ${p.tricksWon}`);
  });

  const human = state.players.find((p) => p.isHuman);
  if (human) {
    lines.push('----------');
    const hand = human.cards
      .map((c, i) => `[${i}]${formatCard(c)}${state.validPlays.includes(i) ? '*' : ''}`)
      .join('  ');
    lines.push(`your hand: ${hand || '(empty)'}`);
    // **脱落しても局は続く。** 打てない理由を名乗らないと固まったように見える。
    if (human.eliminatedAt > 0 && !state.gameEndFlag) {
      lines.push('you are out — the remaining players will finish it between them');
    }
  }

  if (state.gameEndFlag) {
    lines.push('----------');
    lines.push(`game over — ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} held cards longest`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
