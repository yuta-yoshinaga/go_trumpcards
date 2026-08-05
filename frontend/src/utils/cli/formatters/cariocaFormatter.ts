import type { CariocaContractSlot, CariocaResponse } from '../../../types/card';
import {
  formatCard,
  formatCardList,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
} from '../formatterBase';

/** ラウンド終了フェーズ。ここでは手番ではなく次の一手を促す。 */
const ROUND_END_PHASE = 2;

const PHASE_NAMES: Record<number, string> = {
  0: 'DRAW',
  1: 'PLAY',
  2: 'ROUND END',
  3: 'GAME END',
};

/** "trio x3" / "run x4" — mirrors the contract wording of the CUI presenter. */
function slotLabel(slot: CariocaContractSlot): string {
  return `${slot.kind === 0 ? 'trio' : 'run'} x${slot.size}`;
}

/** Format a Carioca game state as terminal text. */
export function formatCariocaState(state: CariocaResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Carioca'));
  lines.push(`round: ${state.roundNumber}/${state.totalRounds}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`contract: ${state.contractSlots.map(slotLabel).join(' + ')}`);
  lines.push(`stock: ${state.drawPileCount} | discard: ${state.discardTop ? formatCard(state.discardTop) : '[  ]'}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const met = p.contractMet ? ' [contract met]' : '';
    lines.push(`${name}: total=${p.cumulativeScore} round=${p.roundScore} cards=${p.cardCount}${met}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
    p.melds.forEach((m, mi) => {
      lines.push(`  [${mi}] ${formatCardList(m.cards)}`);
    });
  }
  lines.push('----------');

  if (state.gameEndFlag) {
    // 手番行は出さない。
  } else if (state.phase === ROUND_END_PHASE) {
    // ラウンド終了時は手番ではなく次の一手を出す (CUI と同じ)。
    lines.push('round over — nr for the next round');
  } else {
    const current = formatPlayerName(state.currentPlayerIdx, state.players[state.currentPlayerIdx]?.isHuman ?? false);
    lines.push(`turn: ${current}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(
      `Game Over! Winner: ${formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false)}`,
    );
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
