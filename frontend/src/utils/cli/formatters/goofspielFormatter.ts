import type { GoofspielResponse } from '../../../types/card';
import { GoofspielPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [GoofspielPhase.BID]: 'BID',
  [GoofspielPhase.REVEAL]: 'REVEAL',
  [GoofspielPhase.GAME_END]: 'GAME END',
};

/** Format a Goofspiel game state as terminal text. */
export function formatGoofspielState(state: GoofspielResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Goofspiel'));
  lines.push(
    `round ${state.roundNumber} | ${state.prizeRemaining} prizes left | ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  // **同時入札であることが規則そのもの。**
  lines.push('everyone bids face down at the same time — highest takes the prize, a tie takes nothing');

  if (state.currentPrize) {
    const carried = state.carriedPrizes.length > 0 ? ` (incl. ${state.carriedPrizes.length} carried)` : '';
    lines.push(`prize: ${formatCard(state.currentPrize)} — ${state.prizeValue} points${carried}`);
  }

  lines.push('----------');

  state.players.forEach((p) => {
    // **伏せたことは見せますが、中身は公開まで見せません。**
    const role = p.revealedBid ? `[played ${formatCard(p.revealedBid)}]` : p.hasBid ? '[has bid]' : '';
    lines.push(`${formatPlayerName(p.id, p.isHuman)}${role}: ${p.cardCount} left, ${p.score} points`);
    // **残り札は全員分を出す。** 使った札は場に出るので隠せていません。
    const hand = p.cards.map((c, i) => (p.isHuman ? `[${i}]${formatCard(c)}` : formatCard(c))).join('  ');
    if (hand) lines.push(`  ${hand}`);
  });

  lines.push('----------');

  if (state.gameEndFlag) {
    const winner = state.players[state.winnerIdx];
    lines.push(
      `game over — ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} took the most (${winner?.score ?? 0})`,
    );
  } else if (state.phase === GoofspielPhase.REVEAL) {
    lines.push(
      state.lastWinnerIdx < 0
        ? 'a tie — nobody takes this prize'
        : `${formatPlayerName(state.lastWinnerIdx, state.lastWinnerIdx === 0)} takes ${state.lastGained} points`,
    );
    lines.push('next — turn the next prize card');
  } else if (state.players.find((p) => p.isHuman)?.hasBid) {
    lines.push('you have bid — waiting for everyone else');
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
