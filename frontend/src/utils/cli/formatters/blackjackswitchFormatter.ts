import type { BlackJackSwitchResponse, Card } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = { 1: 'BET', 2: 'SWITCH', 3: 'ACTION', 4: 'END' };

// Render a row of cards; face-down entries (null) show as "??".
function cardsRow(cards: (Card | null)[]): string {
  return cards.map((c) => (c ? formatCard(c) : '??')).join(' ');
}

/** Format a Blackjack Switch game state as terminal text. */
export function formatBlackjackSwitchState(state: BlackJackSwitchResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Blackjack Switch'));
  lines.push(`chips: ${state.chips} | phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push('----------');

  lines.push(`dealer (${state.dealerScore}): ${cardsRow(state.dealerCards)}`);
  lines.push('----------');

  state.hands.forEach((hand, i) => {
    const marker = state.phase === 3 && i === state.currentHandIdx ? '>' : ' ';
    const flags = [hand.isBJ && 'BJ', hand.busted && 'BUST', hand.stood && 'stood', hand.doubled && 'x2']
      .filter(Boolean)
      .join(',');
    lines.push(
      `${marker}hand${i} (${hand.score}) bet:${hand.bet} ${cardsRow(hand.cards)}${flags ? ` [${flags}]` : ''}`,
    );
  });
  lines.push('----------');

  if (state.phase === 4) {
    lines.push(`total payout: ${state.totalPayout}`);
  }
  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
