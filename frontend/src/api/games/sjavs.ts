// API client for sjavs. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SjavsResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /sjavs/exec endpoint accepts. */
export type SjavsCommand = 'reset' | 'bid' | 'play' | 'next' | 'hint' | 'log';

/**
 * API client for the Sjavs /sjavs/exec endpoint.
 *
 * `bidLength` and `cardIndex` are separate parameters so a bid length can never
 * be read as a hand index. A bid of 0 is a PASS, not an omission.
 */
export const sjavsApi = {
  exec: (command: SjavsCommand, bidLength?: number, cardIndex?: number) =>
    gameExec<SjavsResponse>('sjavs', { command, bidLength, cardIndex }),
};
