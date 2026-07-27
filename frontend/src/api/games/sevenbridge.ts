// API client for sevenbridge. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SevenBridgeResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Seven Bridge game settings. */
export interface SevenBridgeConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** Command verbs accepted by the Seven Bridge /sevenbridge/exec endpoint. */
export type SevenBridgeCommand =
  | 'reset'
  | 'drawstock'
  | 'pon'
  | 'chi'
  | 'meld'
  | 'layoff'
  | 'discard'
  | 'nextround'
  | 'log';

/** API client for the Seven Bridge /sevenbridge/exec endpoint. */
export const sevenBridgeApi = {
  exec: (
    command: SevenBridgeCommand,
    cardIndex?: number,
    config?: SevenBridgeConfigInput,
    cardIndices?: number[],
    targetPlayerIdx?: number,
    meldIdx?: number,
  ) =>
    gameExec<SevenBridgeResponse>('sevenbridge', {
      command,
      cardIndex,
      cardIndices,
      targetPlayerIdx,
      meldIdx,
      config,
    }),
};
