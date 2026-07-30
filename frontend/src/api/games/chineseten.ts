// API client for chineseten. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ChineseTenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /chineseten/exec endpoint accepts. */
export type ChineseTenCommand = 'reset' | 'play' | 'select' | 'hint' | 'log';

/**
 * API client for the Chinese Ten /chineseten/exec endpoint.
 *
 * `play` names a HAND index and `select` a LAYOUT index. They are separate
 * parameters so one can never be read as the other.
 */
export const chinesetenApi = {
  exec: (command: ChineseTenCommand, cardIndex?: number, layoutIndex?: number) =>
    gameExec<ChineseTenResponse>('chineseten', { command, cardIndex, layoutIndex }),
};
