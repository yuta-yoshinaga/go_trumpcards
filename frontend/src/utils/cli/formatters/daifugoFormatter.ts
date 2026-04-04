import type { DaifugoResponse } from '../../../types/card';
import { formatCardList, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

/** Format a Daifugo game state as terminal text. */
export function formatDaifugoState(state: DaifugoResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Daifugo'));

  // Players
  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    if (p.isFinished) {
      lines.push(`${name} finished (rank: ${p.rank})`);
    } else {
      lines.push(`${name} ${p.cardCount} cards`);
    }
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }

  lines.push('----------');

  // Active rules
  const rules: string[] = [];
  if (state.revolutionActive) rules.push('Revolution');
  if (state.elevenBackActive) rules.push('11-Back');
  if (state.suitLocked) rules.push(`Suit-Lock: ${state.lockedSuit}`);
  if (state.tableIsSequence) rules.push('Sequence');
  if (state.reverseDirection) rules.push('Reverse');
  if (state.numberLocked) rules.push('Number-Lock');
  if (state.sequenceLocked) rules.push('Sequence-Lock');
  if (rules.length > 0) lines.push(`[${rules.join(', ')}]`);

  // Exchange actions
  if (state.exchangeActions.length > 0) {
    lines.push('Exchange:');
    for (const ex of state.exchangeActions) {
      const from = formatPlayerName(ex.fromPlayerIdx, state.players[ex.fromPlayerIdx]?.isHuman ?? false);
      const to = formatPlayerName(ex.toPlayerIdx, state.players[ex.toPlayerIdx]?.isHuman ?? false);
      lines.push(`  ${from} \u2192 ${to}: ${formatCardList(ex.cards)}`);
    }
  }

  // Table
  if (state.tableCards.length > 0) {
    const lastPlayer = formatPlayerName(
      state.lastPlayPlayerIdx,
      state.players[state.lastPlayPlayerIdx]?.isHuman ?? false,
    );
    lines.push(`table: ${formatCardList(state.tableCards)} (by ${lastPlayer})`);
  } else {
    lines.push('table: empty (anyone can play)');
  }

  // CPU actions
  if (state.cpuActions.length > 0) {
    for (const a of state.cpuActions) {
      const name = formatPlayerName(a.playerIdx, false);
      if (a.cards.length === 0) {
        lines.push(`${name} passed`);
      } else {
        lines.push(`${name} played ${formatCardList(a.cards)}`);
      }
    }
  }

  // Human action
  if (state.humanAction) {
    const a = state.humanAction;
    if (a.cards.length === 0) {
      lines.push('You passed');
    } else {
      lines.push(`You played ${formatCardList(a.cards)}`);
    }
  }

  // Pending action
  if (state.pendingAction !== 'none') {
    const pending: Record<string, string> = {
      sevenPass: '7-Pass: select cards to pass',
      tenDiscard: '10-Discard: select cards to discard',
      queenBomber: '12-Bomber: select cards to remove',
    };
    lines.push(pending[state.pendingAction] ?? state.pendingAction);
  }

  // Turn info
  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentTurn, state.players[state.currentTurn]?.isHuman ?? false);
    lines.push(`turn: ${current}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push('Game Over');

  lines.push(formatSeparator());
  return lines.join('\n');
}
