import type { KarnoffelResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Play', 'HandEnd', 'GameEnd'];

const SUIT_NAMES: Readonly<Record<number, string>> = { 1: 'S', 2: 'C', 3: 'H', 4: 'D' };

/** Format a Karnöffel game state as terminal text. */
export function formatKarnoffelState(state: KarnoffelResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Karnöffel'));
  lines.push(
    `hand: ${state.handNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}  ` +
      `first to ${state.targetHands}, a hand takes ${state.tricksToWin} tricks`,
  );
  // **切札は表向きの 4 枚のうち最も低い札が決める。**
  lines.push(`chosen suit: ${SUIT_NAMES[state.chosenSuit] ?? '-'} (the LOWEST face-up card picked it)`);
  lines.push(
    `team0: ${state.handsWon[0]} hands (${state.teamTricks[0]} tricks now) | ` +
      `team1: ${state.handsWon[1]} hands (${state.teamTricks[1]} tricks now)`,
  );
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    const hand =
      p.cards.length > 0 ? p.cards.map((c, j) => `[${j}]${formatCard(c)}`).join(' ') : `hidden (${p.cardCount})`;
    const dealer = p.isDealer ? ' [dealer]' : '';
    const turn = p.isCurrentTurn && !state.gameEndFlag ? ' <- turn' : '';
    const up = p.upCard ? formatCard(p.upCard) : '-';
    lines.push(`${name}(T${p.team})${dealer}: up ${up}  tricks ${p.tricksWon}${turn}`);
    lines.push(`  ${hand}`);
  });
  lines.push('----------');

  if (state.trick.length > 0) lines.push(`trick: ${state.trick.map(formatCard).join(' ')}`);

  if (state.phase === 0 && state.currentPlayerIdx === 0 && !state.gameEndFlag) {
    // **悪魔だけ位置が特殊。**表で見せないと「なぜ負けたのか」が分からない。
    lines.push(
      'ranking: J (Karnöffel) > 7 (devil, ONLY WHEN LED) > 6 (Pope) > 2 (Kaiser) > 3 > 4 > 5 > K > Q > 10 > 9 > 8',
    );
    lines.push('  the 3 loses to kings, the 4 to kings and queens, the 5 to every face card');
    lines.push(`playable: ${state.validPlays.join(' ') || '-'}`);
    lines.push('(your turn — play with "p <i>"; no need to follow suit, but the devil cannot lead the first trick)');
  } else if (state.phase === 1 && state.lastResult) {
    const r = state.lastResult;
    lines.push(
      r.winnerTeam >= 0
        ? `(team ${r.winnerTeam} took the hand ${r.tricks[0]}-${r.tricks[1]})`
        : `(neither side reached three tricks ${r.tricks[0]}-${r.tricks[1]})`,
    );
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerTeam >= 0) {
    lines.push(`Game over! Winning team: ${state.winnerTeam}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
