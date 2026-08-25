import type { DilotiResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Diloti game state as terminal text. */
export function formatDilotiState(state: DilotiResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Diloti'));
  lines.push(`round: ${state.roundNumber}  phase: ${state.phase}  target: ${state.config.targetScore}`);
  lines.push(`stock: ${state.deckRemaining}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDealer ? 'Dealer' : 'Player';
    lines.push(`${name} (${role}): taken=${p.capturedCount} xeri=${p.xeri} score=${p.score}`);
    if (p.cards.length > 0) {
      lines.push(`  ${p.cards.map((c, i) => `[${i}]${formatCard(c)}`).join('  ')}`);
    }
  }
  lines.push('----------');

  // **場札にも宣言にも番号が要る。** 取る対象はこの番号で指す。
  if (state.table.length > 0) {
    lines.push(`table: ${state.table.map((c, i) => `[${i}]${formatCard(c)}`).join('  ')}`);
  } else {
    lines.push('table: (empty)');
  }
  state.declarations.forEach((d, i) => {
    const kind = d.isGroup ? 'group, cannot be raised' : 'plain';
    const groups = d.groups.map((g) => g.map(formatCard).join('+')).join(' | ');
    lines.push(`decl[${i}] ${d.value} (${kind}, ${formatPlayerName(d.ownerIdx, d.ownerIdx === 0)}): ${groups}`);
  });

  // **どの札で何ができるかを出す。** 同ランク・合計一致・宣言が混ざるので、
  // 出さないと端末から総当たりで探すことになる。
  state.takeOptions.forEach((opts, i) => {
    if (opts.length === 0) return;
    const rendered = opts
      .map((o) => `(${[...o.tableIdxs.map(String), ...o.declIdxs.map((d) => `d${d}`)].join(' ')})`)
      .join(' ');
    lines.push(`  [${i}] can take: ${rendered}`);
  });
  state.declareOptions.forEach((cands, i) => {
    if (cands.length === 0) return;
    lines.push(`  [${i}] can declare: ${cands.map((c) => `${c.value}:(${c.tableIdxs.join(' ')})`).join(' ')}`);
  });

  if (state.lastResult) {
    for (const line of state.lastResult.lines) {
      if (line.points[0] === 0 && line.points[1] === 0) continue;
      lines.push(`  ${line.key}: ${line.points[0]} - ${line.points[1]}`);
    }
    lines.push(`round total: ${state.lastResult.totals[0]} - ${state.lastResult.totals[1]}`);
  }

  if (state.hintHandIdx >= 0 && isRequestedHint(state)) {
    lines.push(`HINT: play [${state.hintHandIdx}] and ${state.hintAction} (${state.hintReason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(`Game Over! ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} wins!`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
