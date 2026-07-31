import type { SixBidSoloResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Declare', 'Play', 'HandEnd', 'GameEnd'];

/** The six bids in ascending order, with pass at index 0. */
const BID_NAMES = ['pass', 'Solo', 'Heart Solo', 'Misere', 'Guarantee Solo', 'Spread Misere', 'Call Solo'];

const SUIT_NAMES: Readonly<Record<number, string>> = { 1: 'S', 2: 'C', 3: 'H', 4: 'D' };

/** Format a Six-Bid Solo game state as terminal text. */
export function formatSixBidSoloState(state: SixBidSoloResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Six-Bid Solo'));
  lines.push(
    `hand: ${state.handNumber}/${state.targetHands}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}  ` +
      `card points in play: ${state.totalPoints}`,
  );
  if (state.highBid) {
    const trump = state.declared ? (SUIT_NAMES[state.trumpSuit] ?? 'none') : 'not yet named';
    const target = state.bidTargets[state.highBid.kind] ?? 0;
    lines.push(`contract: ${BID_NAMES[state.highBid.kind] ?? '?'}, trump ${trump}, target ${target}`);
    if (state.calledCard) {
      lines.push(`called card: ${formatCard(state.calledCard)} (its holder had to exchange it)`);
    }
  }
  // **ウィドウは精算まで伏せたまま。**枚数だけ出す。
  lines.push(
    state.widow.length > 0 ? `widow: ${state.widow.map(formatCard).join(' ')}` : `widow: ${state.widowSize} face down`,
  );
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    const hand =
      p.cards.length > 0 ? p.cards.map((c, j) => `[${j}]${formatCard(c)}`).join(' ') : `hidden (${p.cardCount})`;
    const dealer = p.isDealer ? ' [dealer]' : '';
    const declarer = p.isDeclarer ? ' [declarer]' : '';
    const turn = p.isCurrentTurn && !state.gameEndFlag ? ' <- turn' : '';
    lines.push(`${name}${dealer}${declarer}: points ${p.points} tricks ${p.tricksWon} total ${p.score}${turn}`);
    lines.push(`  ${hand}`);
  });
  lines.push('----------');

  if (state.trick.length > 0) lines.push(`trick: ${state.trick.map(formatCard).join(' ')}`);

  if (state.phase === 0 && state.bidPlayerIdx === 0 && !state.gameEndFlag) {
    const ladder = BID_NAMES.slice(1)
      .map((n, i) => `${i + 1}:${n}`)
      .join(' < ');
    lines.push(`bids: ${ladder}`);
    lines.push(
      `(your bid — "b <1-6>"; a plain bid must EXCEED ${state.baseTarget}, so ${state.baseTarget + 1} points)`,
    );
  } else if (state.phase === 1 && state.declarerIdx === 0 && !state.gameEndFlag) {
    lines.push('(name trump with "d <1-4>"; a call solo continues "d <trump> <called suit> <called rank>")');
  } else if (state.phase === 2 && state.currentPlayerIdx === 0 && !state.gameEndFlag) {
    lines.push(`playable: ${state.validPlays.join(' ') || '-'}`);
    lines.push('(your turn — play with "p <i>")');
  } else if (state.phase === 3 && state.lastResult) {
    const r = state.lastResult;
    const name = BID_NAMES[r.kind] ?? '?';
    lines.push(
      r.made
        ? `(${name} made: ${r.declarerPoints} points, needed ${r.target})`
        : `(${name} SET: ${r.declarerPoints} points, needed ${r.target})`,
    );
    // **ウィドウは宣言者に入る。ミゼール系だけは 0。**
    lines.push(`widow credited: ${r.widowPoints} (excluded at either misere)`);
    lines.push(`settlement: ${r.value} per opponent  deltas ${r.deltas.join(' / ')}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerIdx >= 0) {
    lines.push(`Game over! Winning seat: ${state.winnerIdx}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
