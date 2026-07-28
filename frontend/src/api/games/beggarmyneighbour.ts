// API client for beggarmyneighbour. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BeggarMyNeighbourResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Beggar-My-Neighbour /beggarmyneighbour/exec endpoint. */
export const beggarmyneighbourApi = {
  exec: (command: 'reset' | 'step' | 'autoplay' | 'log', config?: { maxRounds?: number }) =>
    gameExec<BeggarMyNeighbourResponse>('beggarmyneighbour', { command, ...config }),
};
