// API client for klaberjass. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KlaberjassResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Klaberjass game settings. */
export interface KlaberjassConfigInput {
  cpuDifficulty?: number;
  /** Points needed to win (100-1000, default 501). */
  targetScore?: number;
  /** Whether the schmeiss option is offered (default true). */
  allowSchmeiss?: boolean;
}

/** Commands the /klaberjass/exec endpoint accepts. */
export type KlaberjassCommand =
  | 'reset'
  | 'accept'
  | 'call'
  | 'pass'
  | 'schmeiss'
  | 'answerschmeiss'
  | 'play'
  | 'next'
  | 'log';

/** Options carried alongside a Klaberjass command. */
export interface KlaberjassExecOptions {
  /** Hand index. Required for `play`. */
  cardIndex?: number;
  /** Trump suit for `call`: 1=Spade, 2=Clover, 3=Heart, 4=Diamond. */
  suit?: number;
  /** Required for `answerschmeiss`. `false` makes the thrower the maker. */
  accept?: boolean;
  config?: KlaberjassConfigInput;
}

/**
 * API client for the Klaberjass /klaberjass/exec endpoint.
 *
 * Klaberjass is the two-player ancestor of the Jass family. `accept` takes the
 * turn-up suit as trump and `call` names any other; `schmeiss` offers to throw
 * the deal in, and a refusal makes the *thrower* the maker. During play the
 * server sends `validPlays`, because following suit, trumping when void and
 * overtrumping a trump lead are all compulsory. Only `reset` takes a `config`.
 */
export const klaberjassApi = {
  exec: (command: KlaberjassCommand, opts?: KlaberjassExecOptions) =>
    gameExec<KlaberjassResponse>('klaberjass', {
      command,
      cardIndex: opts?.cardIndex,
      suit: opts?.suit,
      accept: opts?.accept,
      config: opts?.config,
    }),
};
