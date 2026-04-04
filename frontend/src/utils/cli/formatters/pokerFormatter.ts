import type { PokerResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const ACTION_NAMES: Record<number, string> = {
  0: 'Fold',
  1: 'Check',
  2: 'Call',
  3: 'Bet',
  4: 'Raise',
  5: 'All-in',
};

const PHASE_NAMES: Record<number, string> = {
  0: 'ANTE',
  1: 'BET',
  2: 'EXCHANGE',
  3: 'BET2',
  4: 'SHOWDOWN',
  5: 'END',
};

/** Format a Poker game state as terminal text. */
export function formatPokerState(state: PokerResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader(`Poker${state.isLowball ? ' [2-7 Lowball]' : ''}`));
  lines.push(`pot: ${state.pot}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  if (state.jokerCount > 0) lines.push(`jokers: ${state.jokerCount}`);
  lines.push('');

  // Players
  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const status: string[] = [];
    if (p.folded) status.push('Folded');
    if (p.allIn) status.push('All-in');
    const statusStr = status.length > 0 ? ` [${status.join(', ')}]` : '';
    const betStr = p.currentBet > 0 ? ` bet=${p.currentBet}` : '';
    const exchStr = p.exchangeCount >= 0 ? ` exchange=${p.exchangeCount}` : '';
    lines.push(`${name} chips=${p.chips}${betStr}${exchStr}${statusStr}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
    if (state.gameEndFlag && !p.isHuman && p.cards.length > 0) {
      lines.push(`  ${p.cards.map(formatCard).join('  ')}`);
    }
    if (state.gameEndFlag && p.handName) {
      lines.push(`  ${p.handName}`);
    }
  }

  // CPU actions
  if (state.cpuActions.length > 0) {
    lines.push('----------');
    for (const a of state.cpuActions) {
      const name = formatPlayerName(a.playerIdx, false);
      const action = ACTION_NAMES[a.action] ?? 'Unknown';
      const amountStr = a.amount > 0 ? ` (${a.amount})` : '';
      lines.push(`${name}: ${action}${amountStr}`);
    }
  }

  // CPU exchanges
  if (state.cpuExchanges.length > 0) {
    lines.push('----------');
    for (const ex of state.cpuExchanges) {
      const name = formatPlayerName(ex.playerIdx, false);
      lines.push(`${name}: ${ex.exchangeCount} cards exchanged`);
    }
  }

  // Round results
  if (state.roundResults.length > 0) {
    lines.push('----------');
    lines.push('Results:');
    for (const r of state.roundResults) {
      const name = formatPlayerName(r.playerIdx, state.players[r.playerIdx]?.isHuman ?? false);
      lines.push(`  ${name}: ${r.handName} \u2192 ${r.chipsWon > 0 ? '+' : ''}${r.chipsWon}`);
    }
  }

  // Message
  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push('Game Over');

  lines.push(formatSeparator());
  return lines.join('\n');
}
