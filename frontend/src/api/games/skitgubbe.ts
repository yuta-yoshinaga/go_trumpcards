// API client for skitgubbe. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SkitgubbeResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /skitgubbe/exec endpoint accepts. */
export type SkitgubbeCommand = 'reset' | 'play' | 'pickup' | 'hint' | 'log';

/**
 * API client for the Skitgubbe /skitgubbe/exec endpoint.
 *
 * `pickup` takes no index: the server refuses it whenever anything in hand
 * still beats the pile, so there is nothing for the client to choose.
 */
export const skitgubbeApi = {
  exec: (command: SkitgubbeCommand, cardIndex?: number) =>
    gameExec<SkitgubbeResponse>('skitgubbe', { command, cardIndex }),
};
