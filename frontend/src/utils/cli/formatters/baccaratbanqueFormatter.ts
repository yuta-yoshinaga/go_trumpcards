import type { BaccaratBanquePlayer, BaccaratBanqueResponse } from '../../../types/card';
import { BACCARAT_BANQUE_PHASE } from '../../../types/games/baccaratbanque';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<string, string> = {
  [BACCARAT_BANQUE_PHASE.PUNTERS]: 'PUNTERS',
  [BACCARAT_BANQUE_PHASE.BANKER]: 'BANKER DRAW',
  [BACCARAT_BANQUE_PHASE.RESULT]: 'SETTLEMENT',
  [BACCARAT_BANQUE_PHASE.GAME_END]: 'BANK ENDED',
};

const ROLE_NAMES: Record<string, string> = {
  banker: 'Banker (you)',
  right: 'Right tableau',
  left: 'Left tableau',
};

const OUTCOME_NAMES: Record<string, string> = {
  bankerWin: 'banker wins',
  punterWin: 'punter wins',
  tie: 'tie',
};

/** Renders one seat's face-up cards, mod-10 total and chips. */
function seatLine(p: BaccaratBanquePlayer): string {
  const cards = p.cards.length === 0 ? '—' : p.cards.map(formatCard).join(' ');
  const natural = p.cards.length === 2 && p.total >= 8 ? ' *natural' : '';
  const stake = p.bet > 0 ? ` stake ${p.bet}` : '';
  return `${ROLE_NAMES[p.role] ?? p.role}: ${cards} = ${p.total}${natural}  chips ${p.chips}${stake}`;
}

/** Format a Baccarat Banque game state as terminal text. */
export function formatBaccaratBanqueState(state: BaccaratBanqueResponse): string {
  const lines: string[] = [formatHeader('Baccarat Banque')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Coup: ${state.coupNumber}`);
  // **バンクが何回続いているかを出す。** 負けても席が動かないのがこの形式の
  // 要で、残高からは読み取れない。
  lines.push(`Held this bank: ${state.bankHeld} coup(s) — a loss does not end it`);
  lines.push(`Shoe: ${state.shoeRemaining} card(s) left`);

  lines.push(formatSeparator());
  for (const p of state.players) lines.push(seatLine(p));

  if (state.lastResult) {
    lines.push(formatSeparator());
    lines.push(`Settlement (banker ${state.lastResult.bankerTotal}):`);
    // **左右は別勘定。** 1 行にまとめると、片方に払いもう片方から取ったクーが
    // 差額だけになって読めない。
    for (const s of state.lastResult.sides) {
      const role = s.seatIdx === 1 ? 'right' : 'left';
      lines.push(`  ${ROLE_NAMES[role]}: ${OUTCOME_NAMES[s.outcome] ?? s.outcome} (${s.delta})`);
    }
    const net = state.lastResult.bankerDelta;
    lines.push(`  Banker net: ${net > 0 ? `+${net}` : `${net}`}`);
  }

  if (state.gameEndFlag) {
    lines.push(state.retired ? 'You gave up the bank.' : 'The bank has ended.');
  }

  return lines.join('\n');
}
