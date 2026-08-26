import type { CostlyColoursResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Costly Colours game state as terminal text. */
export function formatCostlyColoursState(state: CostlyColoursResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Costly Colours'));
  lines.push(`deal: ${state.dealNumber}  phase: ${state.phase}  target: ${state.config.targetScore}`);
  // **表の 1 枚は常に見せる。** ショーの色役も J / 2 の 4 点もこれ次第。
  lines.push(`turn-up: ${state.turnUp ? formatCard(state.turnUp) : '(none)'}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDealer ? 'Dealer' : 'Elder';
    lines.push(`${name} (${role}): ${p.cardCount} card(s) score=${p.score}`);
    if (p.cards.length > 0) {
      lines.push(`  ${p.cards.map((c, i) => `[${i}]${formatCard(c)}`).join('  ')}`);
    }
    if (p.played.length > 0) {
      lines.push(`  played: ${p.played.map(formatCard).join(' ')}`);
    }
  }
  lines.push('----------');

  const pile = state.pile.length > 0 ? state.pile.map(formatCard).join(' ') : '(nothing yet)';
  lines.push(`count: ${pile}  total ${state.total}`);

  if (state.phase === 'mog') {
    // **断ると相手に 1 点。** 選ぶ前にそう書く。
    lines.push('mog? accepting exchanges a card; refusing pegs 1 for your opponent');
  } else if (state.isHumanTurn) {
    lines.push(
      state.playableIdxs.length > 0
        ? `playable: ${state.playableIdxs.join(' ')}`
        : 'playable: none - nothing fits under 31',
    );
  }

  if (state.lastResult) {
    for (const line of state.lastResult.lines) {
      if (line.points[0] === 0 && line.points[1] === 0) continue;
      lines.push(`  ${line.key}: ${line.points[0]} - ${line.points[1]}`);
    }
    // **どの色役が付いたのかを名指す。** 点だけだと梯子のどこか分からない。
    state.lastResult.combos.forEach((combo, i) => {
      if (!combo) return;
      lines.push(`  ${formatPlayerName(i, i === 0)}: ${combo}`);
    });
    lines.push(`deal total: ${state.lastResult.totals[0]} - ${state.lastResult.totals[1]}`);
  }

  if (state.hintReason && state.hintReason !== 'none' && isRequestedHint(state)) {
    const where = state.hintHandIdx >= 0 ? ` [${state.hintHandIdx}]` : '';
    lines.push(`HINT:${where} (${state.hintReason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(`Game Over! ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} wins!`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
