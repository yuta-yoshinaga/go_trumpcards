import type { ChemindeFerResponse } from '../../../types/card';
import { CHEMIN_DE_FER_RESULT } from '../../../types/games/chemindefer';
import { ChemindeFerPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [ChemindeFerPhase.STAKE]: 'STAKE',
  [ChemindeFerPhase.BET]: 'BET',
  [ChemindeFerPhase.PUNTER_DRAW]: 'PUNTER DRAW',
  [ChemindeFerPhase.BANKER_DRAW]: 'BANKER DRAW',
  [ChemindeFerPhase.ROUND_END]: 'ROUND END',
};

const RESULT_NAMES: Record<number, string> = {
  [CHEMIN_DE_FER_RESULT.none]: 'undecided',
  [CHEMIN_DE_FER_RESULT.banker]: 'banker wins',
  [CHEMIN_DE_FER_RESULT.punter]: 'punters win',
  [CHEMIN_DE_FER_RESULT.tie]: 'a tie',
};

/** Renders a hand and its mod-10 total, or a dash before the deal. */
function handLine(label: string, cards: ChemindeFerResponse['bankerHand'], total: number): string {
  if (cards.length === 0) return `${label}: —`;
  return `${label}: ${cards.map(formatCard).join(' ')} = ${total}`;
}

/** Format a Chemin de Fer game state as terminal text. */
export function formatChemindeFerState(state: ChemindeFerResponse): string {
  const lines: string[] = [formatHeader('Chemin de Fer')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Coup: ${state.roundNumber}${state.config ? ` / ${state.config.rounds}` : ''}`);
  lines.push(`Banker: seat ${state.bankerIdx} (bank ${state.stake})`);

  if (state.stake > 0) {
    lines.push(`Bet total: ${state.totalBet} (uncovered ${state.remainingStake})`);
    if (state.betTurn >= 0) lines.push(`Seat ${state.betTurn} to bet (up to ${state.betMax})`);
    if (state.representativeIdx >= 0) lines.push(`Punter representative: seat ${state.representativeIdx}`);
  }

  if (state.bankerHand.length > 0 || state.punterHand.length > 0) {
    lines.push(formatSeparator());
    lines.push(handLine('Punter', state.punterHand, state.punterTotal));
    lines.push(handLine('Banker', state.bankerHand, state.bankerTotal));
    // **Say when there was no choice.** Otherwise a forced draw reads as the
    // server having decided something the player could have influenced.
    if (state.phase === ChemindeFerPhase.PUNTER_DRAW && !state.punterMayChoose) {
      lines.push('(no choice at this total: 0-4 must draw, 6-7 must stand)');
    }
  }

  lines.push(formatSeparator());
  lines.push(
    `Chips: ${state.players
      .map((p) => `#${p.id}${p.isBanker ? '*' : ''}:${p.chips}${p.bet > 0 ? `(${p.bet})` : ''}`)
      .join(' ')}`,
  );

  if (state.result !== CHEMIN_DE_FER_RESULT.none) {
    lines.push(`Result: ${RESULT_NAMES[state.result] ?? '?'}`);
    // **卓の結果と自分の損益は別の情報** (#5774)。
    const net = state.players.find((p) => p.isHuman)?.lastNet ?? 0;
    lines.push(`your result: ${net > 0 ? `+${net}` : net < 0 ? `${net}` : 'no change'}`);
  }
  if (state.gameEndFlag) lines.push('Game over.');

  return lines.join('\n');
}
