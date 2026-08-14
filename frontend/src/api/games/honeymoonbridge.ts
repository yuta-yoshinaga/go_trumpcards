// API client for honeymoonbridge. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { HoneymoonBridgeConfig, HoneymoonBridgeResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Honeymoon Bridge /honeymoonbridge/exec endpoint.
 *
 * **`bid` needs both level and suit, and `suit: 0` means no-trump.** Omitting
 * the suit is rejected rather than defaulted, because defaulting it would buy
 * a no-trump contract nobody named.
 */
export const honeymoonbridgeApi = {
  exec: (
    command: 'reset' | 'bid' | 'pass' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<HoneymoonBridgeConfig>,
    level?: number,
    suit?: number,
  ) => gameExec<HoneymoonBridgeResponse>('honeymoonbridge', { command, cardIndex, config, level, suit }),
};
