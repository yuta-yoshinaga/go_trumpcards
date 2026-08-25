// API client for cirulla. Follows the split-out convention of gameApi.ts
// (issue #4434); gameApi.ts re-exports this file.

import type { CirullaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Cirulla. */
export interface CirullaConfigInput {
  cpuDifficulty?: number;
  /** Points needed to win, 11-51. */
  targetScore?: number;
}

/** Commands accepted by the /cirulla/exec endpoint. */
export type CirullaCommand = 'reset' | 'play' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Cirulla /cirulla/exec endpoint.
 *
 * **The captured cards ride with the card played.** Sending them separately
 * would leave a board where a card has been played but nothing taken; omit
 * `captureIndices` to lay off, which the server refuses when a capture is
 * available.
 */
export const cirullaApi = {
  exec: (
    command: CirullaCommand,
    opts?: {
      handIndex?: number;
      captureIndices?: number[];
      config?: CirullaConfigInput;
    },
  ) =>
    gameExec<CirullaResponse>('cirulla', {
      command,
      handIndex: opts?.handIndex,
      captureIndices: opts?.captureIndices,
      config: opts?.config,
    }),
};
