// API client for coinche. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CoincheResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Coinche game configuration input shape. */
export interface CoincheConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
  dixDeDer?: number;
  enableBeloteRebelote?: boolean;
}

/** Commands accepted by the Coinche /coinche/exec endpoint. */
export type CoincheCommand =
  | 'reset'
  | 'bid'
  | 'pass'
  | 'coinche'
  | 'surcoinche'
  | 'decline'
  | 'play'
  | 'next'
  | 'nextround'
  | 'hint';

/**
 * API client for the Coinche /coinche/exec endpoint.
 *
 * **The positional slots are the shared trick-game contract, not free
 * parameters.** `useTrickGameBase` dispatches every command as
 * `(command, arg1?, arg2?, config?)` and puts the card index in *arg2* —
 * a client that reads the index out of arg1 sends `cardIndex: undefined`
 * and the server rejects the play. So the mapping from slot to field is
 * written out per command rather than inferred from the parameter names.
 *
 * `bid` needs both halves: arg1 is the target and arg2 the trump suit. A
 * contract is the pair, and the server refuses a request carrying only one
 * rather than filling in the other.
 */
export const coincheApi = {
  exec: (command: CoincheCommand, arg1?: number, arg2?: number, config?: CoincheConfigInput) =>
    gameExec<CoincheResponse>('coinche', {
      command,
      cardIndex: command === 'play' ? arg2 : undefined,
      points: command === 'bid' ? arg1 : undefined,
      suit: command === 'bid' ? arg2 : undefined,
      config,
    }),
};
