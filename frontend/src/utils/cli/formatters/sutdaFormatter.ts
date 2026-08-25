import type { SutdaResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Sutda game state as terminal text. */
export function formatSutdaState(state: SutdaResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Sutda'));
  lines.push(`hand: ${state.handNumber}  phase: ${state.phase}`);
  lines.push(`pot: ${state.pot}  current bet: ${state.currentBet}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDealer ? 'Dealer' : 'Player';
    const folded = p.folded ? ' (folded)' : '';
    lines.push(`${name} (${role}): chips=${p.chips} bet=${p.bet}${folded}`);
    // **伏せているうちは自分のぶんだけ。** 相手の役が見えると賭ける意味が無い。
    if (p.cards.length > 0) {
      lines.push(`  ${p.cards.map(formatCard).join(' ')} - ${p.handName}`);
    }
  }
  lines.push('----------');

  if (state.phase === 'bet' && state.isHumanTurn) {
    lines.push(state.callAmount > 0 ? `to call: ${state.callAmount}` : 'nothing to call - you may check');
    if (state.canRaise) {
      lines.push(`raise: ${state.betUnit} per step (${state.maxRaises - state.raiseCount} left)`);
    }
  }

  if (state.lastResult) {
    const names = state.lastResult.winners.map((w) => formatPlayerName(w, w === 0)).join(', ');
    lines.push(`showdown: ${names} takes ${state.lastResult.pot}`);
  }

  if (state.hintAction && isRequestedHint(state)) {
    lines.push(`HINT: ${state.hintAction} (${state.hintReason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(`Game Over! ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} takes the table!`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
