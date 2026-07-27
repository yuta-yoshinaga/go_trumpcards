// API client for sixcardgolf. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SixCardGolfResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Six Card Golf /sixcardgolf/exec endpoint. */
export const sixcardgolfApi = {
  exec: (params: {
    command: string;
    position?: number;
    config?: { playerCount?: number; cpuDifficulty?: number; rounds?: number };
  }) => gameExec<SixCardGolfResponse>('sixcardgolf', params),
};
