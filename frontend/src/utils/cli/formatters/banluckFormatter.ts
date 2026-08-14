import type { BanLuckResponse } from '../../../types/card';
import { BAN_LUCK_RANK } from '../../../types/games/banluck';
import { BanLuckPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [BanLuckPhase.BET]: 'BET',
  [BanLuckPhase.PLAY]: 'PLAY',
  [BanLuckPhase.ROUND_END]: 'ROUND END',
  [BanLuckPhase.GAME_END]: 'GAME END',
};

const RANK_NAMES: Record<number, string> = {
  [BAN_LUCK_RANK.bust]: 'bust',
  [BAN_LUCK_RANK.point]: 'point',
  [BAN_LUCK_RANK.fiveDragon]: 'five dragon',
  [BAN_LUCK_RANK.banLuck]: 'ban luck',
  [BAN_LUCK_RANK.banBan]: 'ban ban',
};

/** Format a Ban Luck game state as terminal text. */
export function formatBanLuckState(state: BanLuckResponse): string {
  const lines: string[] = [formatHeader('Ban Luck')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Round: ${state.roundNumber}`);
  lines.push(`Banker: ${state.seats[state.bankerSeat]?.name ?? '?'}`);
  // **義務は必ず名指しする。** 拒否されたことだけ伝わっても規則は伝わらない。
  if (state.mustHit) lines.push('The banker cannot stand below 15 — you must draw.');

  lines.push(formatSeparator());
  state.seats.forEach((s, i) => {
    const mark = i === state.turnSeat && state.phase === BanLuckPhase.PLAY ? '*' : ' ';
    const role = s.isBanker ? ' (banker)' : '';
    const cards = s.cards.map(formatCard).join(' ');
    const score = s.cards.length > 0 ? ` = ${s.score}` : '';
    lines.push(`${mark}${s.name}${role} chips ${s.chips} / bet ${s.bet} : ${cards}${score}`);
  });

  if (state.phase === BanLuckPhase.ROUND_END || state.phase === BanLuckPhase.GAME_END) {
    lines.push(formatSeparator());
    state.seats.forEach((s) => {
      lines.push(`${s.name}: ${RANK_NAMES[s.rank] ?? '?'} -> ${s.delta}`);
    });
  }
  if (state.gameEndFlag) {
    lines.push(`Winner: ${state.seats[state.winnerSeat]?.name ?? '?'}`);
  }

  return lines.join('\n');
}
