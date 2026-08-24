import type { QuodlibetResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

/** Format a Quodlibet game state as terminal text. */
export function formatQuodlibetState(state: QuodlibetResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Quodlibet'));
  lines.push(`deal: ${state.dealNumber + 1}/${state.totalDeals}  wheel: ${state.roundNumber}  phase: ${state.phase}`);
  if (state.currentContract >= 0) {
    lines.push(`contract: ${state.currentContractName}`);
    if (!state.isShedding) {
      lines.push(`trick: ${state.trickNumber}/${state.trickCount}`);
    }
  }
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDealer ? 'Dealer' : 'Player';
    // **点は罰点。** 「score」と書くと多いほうが良いように読める。
    lines.push(`${name} (${role}): cards=${p.cardCount} tricks=${p.trickCount} penalty=${p.penalty}`);
    if (p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.isShedding) {
    // **四分と小食いは場が違う。** トリックの欄を出しても何も並ばない。
    if (state.stack.length > 0) {
      lines.push(`stack: ${state.stack.map(formatCard).join(' -> ')}`);
    }
    const placed = state.tablePlaced ?? [];
    if (placed.some((row) => row.length > 0)) {
      lines.push(`table: ${placed.map((row, i) => `s${i + 1}=[${row.join(',')}]`).join(' ')}`);
    }
  } else if (state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  if (state.isContractPhase) {
    // **選べるのはこの輪の残りだけ。** 全 12 種目を並べると嘘になる。
    lines.push(
      `choose: ${state.availableContracts.map((c, i) => `[${c}]${state.availableContractNames[i]}`).join('  ')}`,
    );
  }
  if (state.canPass) lines.push('nothing playable: use pass');

  if (state.lastDeal) {
    lines.push(`last deal (${state.lastDeal.contractName}): ${state.lastDeal.points.join(' / ')}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }
  if (state.hintContract >= 0 && isRequestedHint(state)) {
    lines.push(`HINT: contract [${state.hintContract}]`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    // **勝つのは罰点が最少の人。**
    lines.push(`Game Over! Fewest penalty: ${state.winners.map((w) => formatPlayerName(w, w === 0)).join(', ')}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
