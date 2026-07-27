// API client for trenteetquarante. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TrenteEtQuaranteResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Trente et Quarante (Rouge et Noir) /trenteetquarante/exec endpoint.
 *
 * Trente et Quarante is a pure banking game (no player card decisions).
 *   - `bet` → `(bet, stake)` places the stake on one of the four bets
 *     (0=Noir, 1=Rouge, 2=Couleur, 3=Inverse) and immediately deals both rows
 *     and resolves the round.
 *   - `nextround` starts the next round (chips persist server-side).
 *   - `reset` starts a fresh game (chips persist).
 *   - `log` and `hint` carry no extra fields.
 */
export const trenteetquaranteApi = {
  exec: (command: 'reset' | 'bet' | 'nextround' | 'log' | 'hint', bet?: number, stake?: number) =>
    gameExec<TrenteEtQuaranteResponse>('trenteetquarante', { command, bet, stake }),
};
