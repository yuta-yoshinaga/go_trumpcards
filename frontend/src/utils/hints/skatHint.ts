import type { SkatResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Skat frontend hint stub.
 *
 * The backend already supplies a strategic hint for every interactive Skat phase
 * (bidding, skat pickup, discard, game declaration, and trick play); the
 * frontend renders that server-provided hint via `state.hint`. There is no
 * additional client-side calculation that adds value, so this factory always
 * returns `null` — but the file exists so `useGameHint` can register `skat` in
 * its hintFactories map.
 */
export function getSkatHint(_state: SkatResponse): HintResult | null {
  return null;
}
