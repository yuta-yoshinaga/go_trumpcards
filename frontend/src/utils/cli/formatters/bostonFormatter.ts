import type { BostonResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Bid', 'CallPartner', 'Play', 'HandEnd', 'GameEnd'];

const SUIT_GLYPHS: Record<number, string> = { 1: 'S', 2: 'C', 3: 'H', 4: 'D' };

/** Format a Boston game state as terminal text. */
export function formatBostonState(state: BostonResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Boston'));
  lines.push(`hand: ${state.handNumber}/${state.targetHands}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  if (state.highBid) {
    lines.push(`contract: ${state.highBid.name} trump: ${SUIT_GLYPHS[state.trumpSuit] ?? 'none'}`);
  }
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    const hand =
      p.cards.length > 0 ? p.cards.map((c, j) => `[${j}]${formatCard(c)}`).join(' ') : `hidden (${p.cardCount})`;
    const dealer = p.isDealer ? ' [dealer]' : '';
    const declarer = p.isDeclarer ? ' [declarer]' : '';
    const partner = p.isPartner ? ' [partner]' : '';
    const turn = p.isCurrentTurn && !state.gameEndFlag ? ' <- turn' : '';
    lines.push(`${name}${dealer}${declarer}${partner}: tricks ${p.tricksWon} chips ${p.chips}${turn}`);
    lines.push(`  ${hand}`);
  });
  lines.push('----------');

  if (state.trick.length > 0) lines.push(`trick: ${state.trick.map(formatCard).join(' ')}`);

  if (state.phase === 0 && state.bidPlayerIdx === 0 && !state.gameEndFlag) {
    // **序列を見せないと競りの判断ができない。**ミゼールが間に挟まるため。
    const ladder = state.bidOptions
      .filter((o) => !state.highBid || o.level > state.highBid.level)
      .map((o) => `${o.level}:${o.name}`)
      .join(' < ');
    lines.push(`ladder: ${ladder || '-'}`);
    lines.push('(your bid — "b <step> [suit]" by LADDER STEP, or "ps" to pass)');
  } else if (state.phase === 1 && state.declarerIdx === 0) {
    lines.push('(call a partner with "cp <seat>", or "cp -1" to play alone against three)');
  } else if (state.phase === 2 && state.currentPlayerIdx === 0 && !state.gameEndFlag) {
    lines.push(`playable: ${state.validPlays.join(' ') || '-'}`);
    lines.push('(your turn — play with "p <i>")');
  } else if (state.phase === 3) {
    lines.push(
      state.bidMade
        ? `(contract made with ${state.declarerTricks} tricks — "n" to deal again)`
        : `(contract FAILED with ${state.declarerTricks} tricks)`,
    );
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerIdx >= 0) {
    lines.push(`Game Over! Winner: ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
