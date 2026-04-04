import type { IndianPokerResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'INIT',
  1: 'ANTE',
  2: 'BETTING',
  3: 'SHOWDOWN',
  4: 'END',
};

/** Format an Indian Poker game state as terminal text. */
export function formatIndianpokerState(state: IndianPokerResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Indian Poker'));
  lines.push(`pot: ${state.pot}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}  hand: #${state.handCount}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const status: string[] = [];
    if (p.folded) status.push('Folded');
    if (p.allIn) status.push('All-in');
    const statusStr = status.length > 0 ? ` [${status.join(', ')}]` : '';
    lines.push(`${name} chips=${p.chips}${statusStr}`);
    // In Indian Poker, you see OTHER players' cards but not your own
    if (!p.isHuman && p.card) {
      lines.push(`  ${formatCard(p.card)}`);
    }
    if (p.isHuman) {
      lines.push('  [hidden from you]');
    }
  }
  lines.push('----------');

  if (state.roundResults.length > 0) {
    lines.push('Results:');
    for (const r of state.roundResults) {
      const name = formatPlayerName(r.playerIdx, state.players[r.playerIdx]?.isHuman ?? false);
      const card = r.card ? formatCard(r.card) : '?';
      lines.push(`  ${name}: ${card} ${r.wonAmount > 0 ? `+${r.wonAmount}` : String(r.wonAmount)}`);
    }
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push('Game Over');

  lines.push(formatSeparator());
  return lines.join('\n');
}
