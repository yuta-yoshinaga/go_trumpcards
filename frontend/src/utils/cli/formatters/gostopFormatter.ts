import type { GoStopBreakdown, GoStopResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Play', 'GoDecision', 'RoundEnd', 'GameEnd'];

/** Formats a Go-Stop scoring breakdown as a compact "gwang:N godori:N …" string. */
function formatBreakdown(bd: GoStopBreakdown): string {
  return `gwang:${bd.gwang} godori:${bd.godori} tti:${bd.tti} yeol:${bd.yeol} pi:${bd.pi} = ${bd.goScore}`;
}

/** Format a Go-Stop (ゴーストップ) game state as terminal text. */
export function formatGoStopState(state: GoStopResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Go-Stop'));
  lines.push(
    `round: ${state.roundNumber}  deck: ${state.remainingDeck}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );

  const fieldStr = state.fieldCards.length > 0 ? formatIndexedCards(state.fieldCards) : '(empty)';
  lines.push(`field: ${fieldStr}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const bd = p.breakdown ? ` [${formatBreakdown(p.breakdown)}]` : '';
    lines.push(`${name}: ${p.cardCount} cards  captured=${p.capturedCount}  score=${p.score}  go=${p.goCount}${bd}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.phase === 1 && state.pendingBreakdown) {
    lines.push(`decision: [${formatBreakdown(state.pendingBreakdown)}] = ${state.pendingPoints}  (go / stop)`);
  }

  if (state.phase === 2 && state.lastRoundResult) {
    const d = state.lastRoundResult;
    const winner = d.winner >= 0 ? formatPlayerName(d.winner, state.players[d.winner]?.isHuman ?? false) : 'draw';
    const baks: string[] = [];
    if (d.gwangBak) baks.push('gwang-bak');
    if (d.piBak) baks.push('pi-bak');
    if (d.goBak) baks.push('go-bak');
    const bakStr = baks.length > 0 ? ` (${baks.join(', ')})` : '';
    lines.push(`result: winner=${winner} ${d.basePoints}×${d.goScore}×${d.bakMult}=${d.total}${bakStr}`);
  }

  if (state.hint) {
    if (state.hint.go >= 0) {
      const action = state.hint.go === 1 ? 'go' : 'stop';
      lines.push(`HINT: ${action} (${state.hint.reason})`);
    } else {
      lines.push(`HINT: play ${state.hint.cardIndex} field ${state.hint.fieldIndex} (${state.hint.reason})`);
    }
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
