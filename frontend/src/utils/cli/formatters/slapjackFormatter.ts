import i18n from '../../../i18n';
import type { SlapjackResponse } from '../../../types/card';
import { SlapjackPhase } from '../../../types/phases';

/** Format Slap Jack state for CLI display. Labels are localized via the i18n instance. */
export function formatSlapjackState(s: SlapjackResponse): string {
  const lines: string[] = [];
  const phase =
    s.phase === SlapjackPhase.GAME_END ? i18n.t('slapjack:cli.phase.end') : i18n.t('slapjack:cli.phase.play');
  const top = s.topCard ? `${s.topCard.value}` : '--';
  lines.push(i18n.t('slapjack:cli.summary', { phase, pile: s.centerPileSize, top, turn: s.currentTurnIdx }));
  for (const p of s.players) {
    const tag = p.isHuman ? i18n.t('slapjack:cli.you') : i18n.t('slapjack:cli.cpu');
    lines.push(i18n.t('slapjack:cli.playerStock', { tag, stock: p.stockSize }));
  }
  if (s.message) lines.push(s.message);
  return lines.join('\n');
}
