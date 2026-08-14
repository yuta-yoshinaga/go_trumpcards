// API client for chemindefer. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ChemindeFerResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Chemin de Fer /chemindefer/exec endpoint.
 *
 * **`amount: 0` is a pass, not an omission** — the server rejects a `bet`
 * with no amount, so passing must send the zero explicitly.
 *
 * Drawing and standing are separate commands per side (`pd`/`ps` for the
 * punter, `bd`/`bs` for the banker) so a client cannot resolve the wrong
 * side's decision.
 */
export const chemindeferApi = {
  exec: (
    command:
      | 'reset'
      | 'stake'
      | 'bet'
      | 'pd'
      | 'ps'
      | 'bd'
      | 'bs'
      | 'd'
      | 'st'
      | 'pb'
      | 'next'
      | 'giveup'
      | 'hint'
      | 'log',
    params?: { stake?: number; amount?: number; rounds?: number; chips?: number },
  ) => gameExec<ChemindeFerResponse>('chemindefer', { command, ...params }),
};
