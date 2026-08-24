import type { PiedmonteseTarotResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Scarto', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const OUTCOME_NAMES = ['-', 'Above average', 'Below average'];

/** Format a Tarocco Piemontese game state as terminal text. */
export function formatPiedmonteseTarotState(state: PiedmonteseTarotResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Tarocco Piemontese'));
  lines.push(
    `deal: ${state.roundNumber}  trick: ${state.trickNumber}/${state.trickCount}  ` +
      `phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`scores: ${state.playerScores.map((s, i) => `P${i}=${s}`).join('  ')}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDealer ? 'Dealer' : 'Player';
    lines.push(`${name} (${role}): cards=${p.cardCount} tricks=${p.trickCount} pts=${p.cardPoints} score=${p.score}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trickParts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${trickParts.join(', ')}`);
  }

  // **捨てる枚数を出す。** 卓の大きさで 2 枚と 3 枚が変わるので、書かないと
  // どちらを打てばよいのか端末からは分からない。
  if (state.phase === 0 && state.isHumanScarto) {
    lines.push(`scarto: bury ${state.talonSize} card(s)`);
  }

  if (state.phase === 3 && state.outcome > 0) {
    lines.push(`deal result: ${OUTCOME_NAMES[state.outcome] ?? state.outcome}`);
    lines.push(`deal settlement: ${state.dealScores.map((s, i) => `P${i}=${s > 0 ? `+${s}` : s}`).join('  ')}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(state.winnerPlayer >= 0 ? `Game Over! Winner: Player ${state.winnerPlayer}` : 'Game Over! Draw!');
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
