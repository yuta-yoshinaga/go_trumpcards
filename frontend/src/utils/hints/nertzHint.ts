import i18n from '../../i18n';
import type { NertzResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Human label for one end of a Nertz move. Mirrors `nertzHintZoneLabel` in
 * `internal/adapter/presenter/NertzCuiPresenter.go`, which has been building
 * "Nertz → Foundation 2" from these same fields all along.
 */
function nertzZoneLabel(zone: string, col: number, idx: number): string {
  switch (zone) {
    case 'nertz':
      return i18n.t('nertz:zones.nertz');
    case 'waste':
      return i18n.t('nertz:zones.waste');
    case 'tableau':
      return idx >= 0 ? i18n.t('nertz:zones.tableauWithIdx', { col, idx }) : i18n.t('nertz:zones.tableau', { col });
    case 'foundation':
      return i18n.t('nertz:zones.foundation', { col });
    default:
      return zone;
  }
}

/**
 * Nertz hint stub. Hints are computed server-side and surfaced via
 * `state.hint`; the frontend does not run a separate calculator. Returns
 * a HintResult only when the backend provided one.
 *
 * The reason names the actual move. It used to be the fixed
 * `messages.selectTarget` ("pick a destination"), which threw away the
 * from/to fields the payload carries — the CUI printed them (#4885).
 */
export function getNertzHint(state: NertzResponse): HintResult | null {
  if (!state.hint) return null;
  const { fromZone, fromCol, cardIndex, toZone, toCol } = state.hint;
  const target = `${fromZone}${fromCol >= 0 ? `-c${fromCol}` : ''}${cardIndex >= 0 ? `-i${cardIndex}` : ''}-to-${toZone}-${toCol}`;
  return {
    targetAction: target,
    reason: 'messages.hintMove',
    reasonParams: {
      from: nertzZoneLabel(fromZone, fromCol, cardIndex),
      to: nertzZoneLabel(toZone, toCol, -1),
    },
    confidence: 'moderate',
  };
}
