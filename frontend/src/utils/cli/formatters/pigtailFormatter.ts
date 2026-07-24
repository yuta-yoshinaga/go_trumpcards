import i18n from '../../../i18n';
import type { PigsTailResponse } from '../../../types/card';

/** Format Pig's Tail state for CLI display. Labels are localized via the i18n instance. */
export function formatPigtailState(s: PigsTailResponse): string {
  const lines: string[] = [];
  const phase = s.gameEndFlag ? i18n.t('pigtail:cli.phase.end') : i18n.t('pigtail:cli.phase.play');
  lines.push(i18n.t('pigtail:cli.summary', { phase, circle: s.circleCount, center: s.centerCount }));
  for (const p of s.players) {
    const tag = p.isHuman ? i18n.t('pigtail:cli.you') : i18n.t('pigtail:cli.cpu', { id: p.id });
    lines.push(i18n.t('pigtail:cli.playerCards', { tag, count: p.cardCount }));
  }
  if (s.lastDrawCard) {
    const card = `${s.lastDrawCard.design[0]}${s.lastDrawCard.value}`;
    const status = s.lastPenalty ? i18n.t('pigtail:cli.penalty') : i18n.t('pigtail:cli.safe');
    lines.push(i18n.t('pigtail:cli.last', { card, status }));
  }
  if (s.message) lines.push(s.message);
  return lines.join('\n');
}
