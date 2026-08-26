import type { CirullaResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Cirulla game state as terminal text. */
export function formatCirullaState(state: CirullaResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Cirulla'));
  lines.push(`round: ${state.roundNumber}  phase: ${state.phase}  target: ${state.config.targetScore}`);
  lines.push(`stock: ${state.deckRemaining}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDealer ? 'Dealer' : 'Player';
    lines.push(`${name} (${role}): taken=${p.capturedCount} denari=${p.denariCount} scopa=${p.scope} score=${p.score}`);
    if (p.lastBonus) lines.push(`  bonus: ${p.lastBonus}`);
    if (p.cards.length > 0) {
      lines.push(`  ${p.cards.map((c, i) => `[${i}]${formatCard(c)}`).join('  ')}`);
    }
  }
  lines.push('----------');

  // **場札には番号が要る。** 取る札はこの番号で指す。
  if (state.table.length > 0) {
    lines.push(`table: ${state.table.map((c, i) => `[${i}]${formatCard(c)}`).join('  ')}`);
  } else {
    lines.push('table: (empty)');
  }

  // **どの札で何が取れるかを出す。** 3 つの規則が混ざるので、出さないと
  // 端末から総当たりで探すことになる。
  state.captureOptions.forEach((groups, i) => {
    if (groups.length === 0) return;
    lines.push(`  [${i}] can take: ${groups.map((g) => `(${g.join(' ')})`).join(' ')}`);
  });

  if (state.lastResult) {
    for (const line of state.lastResult.lines) {
      if (line.points[0] === 0 && line.points[1] === 0) continue;
      lines.push(`  ${line.key}: ${line.points[0]} - ${line.points[1]}`);
    }
    lines.push(`round total: ${state.lastResult.totals[0]} - ${state.lastResult.totals[1]}`);
    if (state.lastResult.sweptDenari >= 0) {
      lines.push(
        `all denari to ${formatPlayerName(state.lastResult.sweptDenari, state.lastResult.sweptDenari === 0)}!`,
      );
    }
  }

  if (state.hintHandIdx >= 0 && isRequestedHint(state)) {
    const take = state.hintCaptureIdxs.length > 0 ? `take (${state.hintCaptureIdxs.join(' ')})` : 'lay off';
    lines.push(`HINT: play [${state.hintHandIdx}] and ${take} (${state.hintReason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(`Game Over! ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} wins!`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
