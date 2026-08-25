import type { CometResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

/** Cards of the current sequence to show; older ones scroll off. */
const PILE_TAIL = 8;

/** Format a Comet game state as terminal text. */
export function formatCometState(state: CometResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Comet'));
  lines.push(`round: ${state.roundNumber}  phase: ${state.phase}  target: ${state.config.targetScore}`);
  // **死に手の枚数は見せる。** ここに眠った札で連なりが止まる。
  lines.push(`dead hand: ${state.deadCount}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDealer ? 'Dealer' : 'Player';
    lines.push(`${name} (${role}): ${p.cardCount} card(s) score=${p.score}`);
    if (p.cards.length > 0) {
      lines.push(`  ${p.cards.map((c, i) => `[${i}]${formatCard(c)}`).join('  ')}`);
    }
  }
  lines.push('----------');

  const shown = state.pile.slice(-PILE_TAIL);
  lines.push(shown.length > 0 ? `sequence: ${shown.map(formatCard).join(' -> ')}` : 'sequence: (nothing yet)');
  // **スートは問わない。** 数字だけで昇る。
  lines.push(state.need > 0 ? `need: rank ${state.need} (any suit)` : 'need: lead any card');

  if (state.isHumanTurn) {
    lines.push(
      state.playableIdxs.length > 0 ? `playable: ${state.playableIdxs.join(' ')}` : 'playable: none - you must pass',
    );
  }

  if (state.lastResult) {
    const r = state.lastResult;
    lines.push(`${formatPlayerName(r.winnerIdx, r.winnerIdx === 0)} went out for ${r.gained[r.winnerIdx]}`);
    lines.push(`  kings never played: ${r.unplayedKings}`);
    if (r.heldWildIdx >= 0) {
      lines.push(`  ${formatPlayerName(r.heldWildIdx, r.heldWildIdx === 0)} held the Comet: -1`);
    }
  }

  if (state.hintHandIdx >= 0 && isRequestedHint(state)) {
    lines.push(`HINT: play [${state.hintHandIdx}] (${state.hintReason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(`Game Over! ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} wins!`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
