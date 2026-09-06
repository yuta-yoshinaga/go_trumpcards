// API client for omi. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OmiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Omi game settings. */
export interface OmiConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Omi /omi/exec endpoint.
 * Supported commands match the OmiWebController dispatch table. */
export const omiApi = {
  exec: (
    command: 'reset' | 'calltrump' | 'play' | 'next' | 'nextround' | 'hint',
    cardIndex?: number,
    suit?: number,
    _goAlone?: undefined,
    config?: OmiConfigInput,
  ) =>
    gameExec<OmiResponse>('omi', {
      command,
      cardIndex,
      suit,
      config,
    }),
};
