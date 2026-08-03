import type { LiteratureResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Play', 'GameEnd'];

const SUIT_NAMES: Readonly<Record<number, string>> = { 1: 'S', 2: 'C', 3: 'H', 4: 'D' };

/** 0 = open, 1 = team 0, 2 = team 1, 3 = cancelled. */
const STATE_NAMES = ['open', 'team 0', 'team 1', 'CANCELLED'];

/** Name a half-suit from its index: suit, then low (2-7) or high (9-A). */
function halfSuitName(half: number): string {
  const suit = SUIT_NAMES[1 + Math.floor(half / 2)] ?? '?';
  return `${suit} ${half % 2 === 0 ? 'low (2-7)' : 'high (9-A)'}`;
}

/** Format a Literature game state as terminal text. */
export function formatLiteratureState(state: LiteratureResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Literature'));
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  // **無効は別枠。**合計が 8 になるとは限らない。
  lines.push(
    `team0: ${state.teamHalfSuits[0]} | team1: ${state.teamHalfSuits[1]} | ` +
      `cancelled: ${state.cancelledCount} | open: ${state.openCount}`,
  );
  lines.push(`winning takes ${state.winThreshold} of ${state.halfSuitCnt} — a MAJORITY, so four decides nothing`);
  lines.push('----------');

  state.halfSuits.forEach((st, half) => {
    lines.push(`  [${half}] ${halfSuitName(half)}: ${STATE_NAMES[st] ?? st}`);
  });
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    // **終局まで味方の手札も見えない。**
    const hand = p.cards.length > 0 ? p.cards.map(formatCard).join(' ') : `hidden (${p.cardCount})`;
    const turn = p.isCurrentTurn && !state.gameEndFlag ? ' <- turn' : '';
    lines.push(`seat ${i} ${name}(T${p.team})${turn}: ${p.cardCount} cards  ${hand}`);
  });
  lines.push('----------');

  // **要求の履歴は公開情報。**直近だけ出す。
  if (state.asks.length > 0) {
    lines.push('recent asks (everyone sees these):');
    for (const a of state.asks.slice(-5)) {
      const card = a.card ? formatCard(a.card) : '?';
      lines.push(`  seat ${a.from} -> seat ${a.to}: ${card} ... ${a.success ? 'hit' : 'miss'}`);
    }
  }

  if (state.lastClaim) {
    const c = state.lastClaim;
    const name = halfSuitName(c.halfSuit);
    // **無効は「相手に渡る」とは違う。**
    if (c.outcome === 1) {
      lines.push(`(${name} was CANCELLED — the placement was wrong; it does NOT go to the opponents)`);
    } else if (c.outcome === 2) {
      lines.push(`(${name} went to the opponents, team ${c.awardedTeam} — they held at least one)`);
    } else {
      lines.push(`(${name} went to team ${c.awardedTeam})`);
    }
  }

  if (state.phase === 0 && state.currentPlayerIdx === 0 && !state.gameEndFlag) {
    lines.push('(your turn — "a <seat> <1-4> <1-13>" to ask, "c <half> <seat x6>" to claim)');
    lines.push('  you may ask ONLY an opponent, only for a half-suit you hold, and only for a card you do NOT hold');
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(state.winnerTeam >= 0 ? `Game over! Winning team: ${state.winnerTeam}` : 'Game over! It ended level.');
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
