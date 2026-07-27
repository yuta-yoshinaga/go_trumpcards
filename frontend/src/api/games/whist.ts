// API client for whist. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { WhistConfig, WhistResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Whist /whist/exec endpoint. */
export const whistApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<WhistConfig>,
  ) => gameExec<WhistResponse>('whist', { command, cardIndex, config }),
};
