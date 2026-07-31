import type { KilleResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Exchange', 'Showdown', 'GameEnd'];

/** Why a seat went out. The Hussar and the Pig fire regardless of strength. */
function outReason(knockedBy: string): string {
  if (knockedBy === 'hussar') return 'Hussar';
  if (knockedBy === 'pig') return 'Pig';
  return 'lowest';
}

/** Format a Kille game state as terminal text. */
export function formatKilleState(state: KilleResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Kille'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(
    `dealer: ${formatPlayerName(state.dealerIdx, state.dealerIdx === 0)}  pot: ${state.pot}  stock: ${state.stockCount}`,
  );
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    const card = p.card ? ` ${formatCard(p.card)}` : ' [hidden]';
    const turn = p.isCurrentTurn && !state.gameEndFlag ? ' <- turn' : '';
    const dealer = i === state.dealerIdx ? ' [dealer]' : '';
    const status = p.isFinished ? ' (eliminated)' : p.isOut ? ` (out: ${outReason(p.knockedBy)})` : '';
    const pat = p.isSatisfied && !p.isOut ? ' [stands pat]' : '';
    lines.push(`${name}: chips ${p.chips}${card}${dealer}${pat}${status}${turn}`);
  });
  lines.push('----------');

  for (const e of state.events) {
    const actor = formatPlayerName(e.actor, e.actor === 0);
    const target = e.target >= 0 ? formatPlayerName(e.target, e.target === 0) : 'the stock';
    lines.push(`${e.kind}: ${actor} -> ${target}`);
  }

  if (state.phase === 0 && state.currentPlayerIdx === 0 && !state.gameEndFlag) {
    lines.push('(your turn — exchange with "e" or keep your card with "s")');
  } else if (state.phase === 1 && !state.gameEndFlag) {
    lines.push('(showdown — buy back in with "re", or deal again with "nr")');
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerIdx >= 0) {
    lines.push(`Game Over! Winner: ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
