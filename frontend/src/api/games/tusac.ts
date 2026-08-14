// API client for tusac. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TuSacResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Tu Sac /tusac/exec endpoint.
 *
 * **Drawing from the stock and taking the discard are separate commands**
 * rather than a flag, so a missing field can never silently turn a deliberate
 * pickup into a stock draw. `index` and `indexes` are 0-based hand positions;
 * a card cannot be named, because the deck holds four copies of every
 * colour-and-piece.
 */
export const tusacApi = {
  exec: (
    command: 'reset' | 'draw' | 'take' | 'meld' | 'discard' | 'next' | 'hint' | 'log',
    params?: { index?: number; indexes?: number[] },
  ) => gameExec<TuSacResponse>('tusac', { command, ...params }),
};
