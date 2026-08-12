// API client for stealingbundles. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { StealingBundlesConfig, StealingBundlesResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Stealing Bundles /stealingbundles/exec endpoint.
 *
 * **`steal` names a victim as well as a card.** Taking a bundle is a different
 * action from capturing off the table, and the server rejects it without both.
 */
export const stealingbundlesApi = {
  exec: (
    command: 'reset' | 'take' | 'steal' | 'trail' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    victimIdx?: number,
    config?: Partial<StealingBundlesConfig>,
  ) => gameExec<StealingBundlesResponse>('stealingbundles', { command, cardIndex, victimIdx, config }),
};
