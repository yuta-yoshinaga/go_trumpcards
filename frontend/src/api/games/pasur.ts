// API client for pasur. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PasurConfig, PasurResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Pasur /pasur/exec endpoint.
 *
 * **An absent `table` means "lay the card down", not "parameter missing".**
 * Whether laying down was legal is the server's call — capturing is compulsory
 * when a capture exists.
 */
export const pasurApi = {
  exec: (
    command: 'reset' | 'play' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<PasurConfig>,
    table?: number[],
  ) => gameExec<PasurResponse>('pasur', { command, cardIndex, config, table }),
};
