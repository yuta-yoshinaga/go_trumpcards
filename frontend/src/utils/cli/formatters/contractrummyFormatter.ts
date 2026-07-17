import type { ContractRummyContractSlot, ContractRummyResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = { 0: 'DRAW', 1: 'PLAY', 2: 'ROUND END', 3: 'GAME END' };

/** Describe a contract slot as e.g. "set(3)" or "run(4)". */
function formatSlot(slot: ContractRummyContractSlot): string {
  return `${slot.kind === 0 ? 'set' : 'run'}(${slot.size})`;
}

/** Format a Contract Rummy game state as terminal text. */
export function formatContractRummyState(state: ContractRummyResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Contract Rummy'));
  lines.push(`round ${state.roundNumber}/${state.totalRounds} | phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(`contract: ${state.contractSlots.map(formatSlot).join(' + ')}`);
  const discard = state.discardTop ? formatCard(state.discardTop) : '—';
  lines.push(`stock: ${state.drawPileCount} | discard: ${discard}`);
  lines.push('----------');

  state.players.forEach((p) => {
    const marker = p.id === state.currentPlayerIdx && !state.gameEndFlag ? '>' : ' ';
    const met = p.contractMet ? ' [contract met]' : '';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}: ${p.cardCount} cards | score ${p.cumulativeScore}${met}`,
    );
    p.melds.forEach((meld, mi) => {
      lines.push(`    M${mi}: ${meld.cards.map(formatCard).join(' ')}`);
    });
  });
  lines.push('----------');

  const human = state.players.find((p) => p.isHuman);
  if (human) {
    const hand = human.cards.map((c, i) => `[${i}]${formatCard(c)}`).join('  ');
    lines.push(`your hand: ${hand || '(empty)'}`);
  }

  if (state.gameEndFlag && state.winnerIdx >= 0) {
    lines.push('----------');
    lines.push(
      `game over — winner: ${formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false)}`,
    );
  } else if (state.phase === 2 && state.roundWinnerIdx >= 0) {
    lines.push('----------');
    lines.push(
      `round winner: ${formatPlayerName(state.roundWinnerIdx, state.players[state.roundWinnerIdx]?.isHuman ?? false)}`,
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
