// API client for zwicker. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ZwickerResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /zwicker/exec endpoint accepts. */
export type ZwickerCommand = 'reset' | 'take' | 'build' | 'trail' | 'next' | 'hint' | 'log';

/** Index selections and values a Zwicker command may carry. */
export interface ZwickerParams {
  cardIndex?: number;
  /**
   * Which matching value the played card is used as. Required for `take` —
   * an ace or court card has two, so the card alone does not determine the
   * capture.
   */
  playedValue?: number;
  tableIndices?: number[];
  buildIndices?: number[];
  /** Value of the build being made. Required for `build`. */
  declaredValue?: number;
}

/** API client for the Zwicker /zwicker/exec endpoint. */
export const zwickerApi = {
  exec: (command: ZwickerCommand, params?: ZwickerParams) =>
    gameExec<ZwickerResponse>('zwicker', { command, ...params }),
};
