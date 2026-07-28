// API client for tarneeb. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TarneebResponse } from '../../types/card';
import { gameExec } from '../gameExec';
import { twoTenJackApi } from './twotenjack';

/** Configuration options for Tarneeb game settings. */
export interface TarneebConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
  minBid?: number;
}

/**
 * API client for the Tarneeb /tarneeb/exec endpoint.
 *
 * Signature mirrors {@link twoTenJackApi}: `(command, arg1, cardIndex, config)`.
 * `arg1` is overloaded based on the command:
 *   - `command === 'bid'` → `arg1` is the bid value (0=pass, 7-13=bid).
 *   - `command === 'trump'` → `arg1` is the trump suit (1=♠ 2=♣ 3=♥ 4=♦).
 *   - otherwise `arg1` is ignored.
 */
export const tarneebApi = {
  exec: (
    command: 'reset' | 'bid' | 'trump' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    arg1?: number,
    cardIndex?: number,
    config?: TarneebConfigInput,
  ) => {
    const body: {
      command: string;
      bid?: number;
      trumpSuit?: number;
      cardIndex?: number;
      config?: TarneebConfigInput;
    } = { command };
    if (command === 'bid') body.bid = arg1;
    else if (command === 'trump') body.trumpSuit = arg1;
    if (cardIndex !== undefined) body.cardIndex = cardIndex;
    if (config) body.config = config;
    return gameExec<TarneebResponse>('tarneeb', body);
  },
};
