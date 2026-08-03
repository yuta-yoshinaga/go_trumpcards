import type { RussianBankResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Russian Bank source zones, as the backend numbers them. */
const ZONE_NAMES: Record<number, string> = { 0: 'reserve', 1: 'waste', 2: 'tableau' };

/**
 * Returns a Russian Bank frontend hint derived from the backend hint, or null.
 *
 * The suggestion rides along with every state response (see
 * RussianBankWebPresenter.Output, #4483). A card reaching a foundation is
 * always progress; anything else is a rearrangement that may or may not be.
 */
export function getRussianBankHint(state: RussianBankResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  const zone = ZONE_NAMES[hint.zone];
  if (!zone) return null;

  // **列を持つのはタブローだけ。**リザーブと廃札は 1 山で、バックエンドは
  // `RussianBankSource{Zone: RussianBankZoneReserve}` のように Col を Go の
  // ゼロ値 0 のまま送る。つまり col では区別できないので、ゾーンで判定する。
  // col を無条件に付けると `reserve-0`（存在しない列）になる。
  const target = zone === 'tableau' ? `${zone}-${hint.col}` : zone;
  return {
    targetAction: target,
    reason: hint.toFoundation ? 'frontendHint.russianbankToFoundation' : 'frontendHint.russianbankToTableau',
    confidence: hint.toFoundation ? 'strong' : 'moderate',
  };
}
