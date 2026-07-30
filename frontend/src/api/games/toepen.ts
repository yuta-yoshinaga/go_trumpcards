// API client for toepen. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ToepenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /toepen/exec endpoint accepts. */
export type ToepenCommand = 'reset' | 'play' | 'toep' | 'answer' | 'next' | 'hint' | 'log';

/**
 * API client for the Toepen /toepen/exec endpoint.
 *
 * `answer` carries `stay`: true stays in after a toep, false folds. It is a
 * separate parameter from `cardIndex` so the two can never be confused.
 */
export const toepenApi = {
  exec: (command: ToepenCommand, cardIndex?: number, stay?: boolean) =>
    gameExec<ToepenResponse>('toepen', { command, cardIndex, stay }),
};
