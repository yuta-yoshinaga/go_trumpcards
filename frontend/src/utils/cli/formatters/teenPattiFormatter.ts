import type { TeenPattiResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Betting', 'SideShow', 'Showdown', 'RoundEnd', 'GameEnd'];

/** Format a Teen Patti game state as terminal text. */
export function formatTeenPattiState(state: TeenPattiResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Teen Patti'));
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
  if (state.canRequestSideShow) {
    lines.push('(you may request a Side Show with the previous Seen player)');
  }
  if (state.sideShowTarget >= 0 && state.sideShowRequester >= 0) {
    const requester = formatPlayerName(
      state.sideShowRequester,
      state.players[state.sideShowRequester]?.isHuman ?? false,
    );
    const target = formatPlayerName(state.sideShowTarget, state.players[state.sideShowTarget]?.isHuman ?? false);
    lines.push(`Side Show: ${requester} -> ${target} (accept / decline)`);
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
