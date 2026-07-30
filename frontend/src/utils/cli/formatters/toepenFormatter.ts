import type { ToepenPlayer, ToepenResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

/**
 * Render one seat. The hand prints only when the server sent it; a hidden seat
 * arrives with a count and no cards.
 */
function formatSeat(p: ToepenPlayer, maxLives: number): string {
  const who = p.isHuman ? 'you' : `cpu${p.id}`;
  const status = p.eliminated ? ' [out]' : p.folded ? ' [folded]' : '';
  const hand = p.hidden ? `${p.cardCount} cards` : p.cards.map((c, i) => `${i}:${formatCard(c)}`).join(' ');
  return `${who}: ${p.lives}/${maxLives} lives${status}  ${hand}`;
}

/** Format a Toepen game state as terminal text. */
export function formatToepenState(state: ToepenResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Toepen'));
  lines.push(`hand ${state.handNumber} / trick ${state.trickNumber} · stake ${state.stake}`);
  // The ranking prints every time: it is inverted from every other game here
  // and is the single easiest thing to misplay.
  lines.push('ranking: 10 > 9 > 8 > 7 > A > K > Q > J');
  lines.push('----------');

  for (const p of state.players) {
    lines.push(formatSeat(p, state.maxLives));
  }

  if (state.currentTrick.length > 0) {
    lines.push(`trick: ${state.currentTrick.map((tc) => formatCard(tc.card)).join(' ')}`);
  }
  if (state.knockerIdx >= 0) {
    lines.push(`toep by seat ${state.knockerIdx} -- stake is ${state.stake}; s to stay, f to fold`);
  }

  if (state.gameEndFlag) {
    lines.push(state.winnerIdx === 0 ? 'you win' : 'you lose');
  }

  return lines.join('\n');
}
