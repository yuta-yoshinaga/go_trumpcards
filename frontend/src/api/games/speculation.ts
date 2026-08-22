// API client for Speculation. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SpeculationResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands accepted by `speculationDispatch` (SpeculationWebController.go). */
export type SpeculationCommand = 'reset' | 'flip' | 'accept' | 'decline' | 'bid' | 'next' | 'hint' | 'log';

/**
 * API client for the Speculation /speculation/exec endpoint.
 *
 * **`bid` must carry an `amount`.** The server rejects a `bid` with no amount
 * rather than reading the omission as 0 — a 0 bid and a decline would otherwise
 * be indistinguishable, and the domain only accepts a raise *above* the
 * standing offer anyway.
 */
export const speculationApi = {
  exec: (command: SpeculationCommand, params?: { amount?: number }) =>
    gameExec<SpeculationResponse>('speculation', { command, ...params }),
};
