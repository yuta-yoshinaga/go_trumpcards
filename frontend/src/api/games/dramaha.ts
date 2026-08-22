// API client for dramaha. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { DramahaResponse } from '../../types/card';
import { gameExec } from '../gameExec';
import type { HoldemConfigInput, HoldemLikeCommand } from './holdem';

/**
 * Configuration options for Dramaha (Hold'em's options plus the draw round's
 * card selection).
 *
 * `indices` rides in the config slot rather than as a sixth positional
 * parameter so this exec stays assignable to the shared community-poker exec
 * type. On the wire it lands at the top level of the request body, which is
 * where `controller.DramahaWebInput` reads it.
 */
export interface DramahaConfigInput extends HoldemConfigInput {
  /**
   * **0-based** hole-card positions to exchange in the draw round.
   *
   * Read only by the `draw` command. Omitted or empty means "stand pat" — the
   * backend does not distinguish the two. (The CLI's `d` command takes 1-based
   * numbers to match what is printed on screen and converts before it gets
   * here; see {@link parseDramahaCommand}.)
   */
  indices?: number[];
}

/** Every command the Dramaha endpoint accepts: Hold'em's set plus `draw`. */
export type DramahaCommand = HoldemLikeCommand | 'draw';

/**
 * API client for the Dramaha /dramaha/exec endpoint.
 *
 * Not `createHoldemLikeApi`: Dramaha is the only game in the family with a
 * draw round, and widening the shared factory would put a `draw` command on
 * the API surface of fifteen games that would reject it — the same reason the
 * Go side embeds `HoldemWebInput` instead of adding `indices` to it.
 */
export const dramahaApi = {
  exec: (
    command: DramahaCommand,
    amount?: number,
    config?: DramahaConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<DramahaResponse>('dramaha', {
      command,
      amount,
      humanPlayMs,
      profile,
      ...config,
    }),
};
