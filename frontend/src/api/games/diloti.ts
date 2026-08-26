// API client for diloti. Follows the split-out convention of gameApi.ts
// (issue #4434); gameApi.ts re-exports this file.

import type { DilotiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Diloti. */
export interface DilotiConfigInput {
  cpuDifficulty?: number;
  /** Points needed to win, 21-101. */
  targetScore?: number;
}

/** Commands accepted by the /diloti/exec endpoint. */
export type DilotiCommand = 'reset' | 'play' | 'nextround' | 'hint' | 'log';

/** What to do with the played card. */
export type DilotiAction = 'capture' | 'declare' | 'trail';

/**
 * API client for the Diloti /diloti/exec endpoint.
 *
 * **The capture targets and the declared value ride with the card played.**
 * Sending them separately would leave a board where a card has been played but
 * nothing taken.
 */
export const dilotiApi = {
  exec: (
    command: DilotiCommand,
    opts?: {
      handIndex?: number;
      action?: DilotiAction;
      tableIndices?: number[];
      declIndices?: number[];
      declValue?: number;
      config?: DilotiConfigInput;
    },
  ) =>
    gameExec<DilotiResponse>('diloti', {
      command,
      handIndex: opts?.handIndex,
      action: opts?.action,
      tableIndices: opts?.tableIndices,
      declIndices: opts?.declIndices,
      declValue: opts?.declValue,
      config: opts?.config,
    }),
};
