// API client for botifarra. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BotifarraResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Botifarra /botifarra/exec endpoint.
 *
 * **`suit` must be sent even when it is -1.** No-trump is a real declaration,
 * and suits are 1..4, so the server distinguishes "no trump" from "absent".
 */
export const botifarraApi = {
  exec: (
    command: 'reset' | 'declare' | 'delegate' | 'double' | 'passdouble' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    suit?: number,
  ) => gameExec<BotifarraResponse>('botifarra', { command, cardIndex, suit }),
};
