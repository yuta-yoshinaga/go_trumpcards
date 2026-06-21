import type { ThreeCardBragResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Betting', 'Showdown', 'RoundEnd', 'GameEnd'];

/** Format a Three Card Brag game state as terminal text. */
export function formatThreeCardBragState(state: ThreeCardBragResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Three Card Brag'));
  lines.push(`deal: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(`pot: ${state.pot}  stake: ${state.stake}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const status = p.out ? 'OUT' : p.folded ? 'folded' : p.seen ? 'seen' : 'blind';
    const handText = p.handName ? `  hand=${p.handName}` : '';
    lines.push(`${name}: chips=${p.chips} bet=${p.roundBet} [${status}]${handText}`);
    if (p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.canShow) {
    lines.push('(you may Show to force a showdown)');
  }

  if (state.hint) {
    lines.push(`HINT: ${state.hint.action} (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.matchWinnerIdx >= 0) {
    lines.push(
      `Game Over! Winner: ${formatPlayerName(state.matchWinnerIdx, state.players[state.matchWinnerIdx]?.isHuman ?? false)}`,
    );
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
