// API client for laughandliedown. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { LaughAndLieDownResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /laughandliedown/exec endpoint accepts. */
export type LaughAndLieDownCommand = 'reset' | 'play' | 'hint' | 'log';

/**
 * API client for the Laugh and Lie Down /laughandliedown/exec endpoint.
 *
 * `takeCount` is optional and defaults to 1 server-side: an ordinary single
 * capture needs no extra field, and only 1 or 3 are legal.
 */
export const laughandliedownApi = {
  exec: (command: LaughAndLieDownCommand, cardIndex?: number, takeCount?: number) =>
    gameExec<LaughAndLieDownResponse>('laughandliedown', { command, cardIndex, takeCount }),
};
