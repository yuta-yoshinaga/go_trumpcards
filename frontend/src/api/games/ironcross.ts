// API client for ironcross. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { IronCrossResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Iron Cross /ironcross/exec endpoint.
 *
 * `amount` is required for `bet` and `raise` and ignored by the other actions.
 * The arm of the cross is picked with the named `vertical` / `horizontal`
 * commands, so no body field is needed for the one choice the game has.
 */
export const ironcrossApi = {
  exec: (
    command:
      | 'reset'
      | 'fold'
      | 'check'
      | 'call'
      | 'bet'
      | 'raise'
      | 'vertical'
      | 'horizontal'
      | 'next'
      | 'hint'
      | 'log',
    params?: { amount?: number },
  ) => gameExec<IronCrossResponse>('ironcross', { command, ...params }),
};
