// API client for karnoffel. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KarnoffelResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Karnöffel game settings. */
export interface KarnoffelConfigInput {
  cpuDifficulty?: number;
  /** Hands needed to win the game. */
  targetHands?: number;
}

/** Commands the /karnoffel/exec endpoint accepts. */
export type KarnoffelCommand = 'reset' | 'play' | 'next' | 'log';

/** Options carried alongside a Karnöffel command. */
export interface KarnoffelExecOptions {
  /** Hand index. Required for `play`. */
  cardIndex?: number;
  config?: KarnoffelConfigInput;
}

/**
 * API client for the Karnöffel /karnoffel/exec endpoint.
 *
 * Karnöffel deals **five cards each** from a 48-card pack with the aces
 * removed; the first card of each hand is dealt face up and **the lowest of
 * those four picks the chosen suit**. Within it the **jack is the Karnöffel**
 * and beats everything, the **seven is the devil and only wins when led**, and
 * the 3/4/5 are partial trumps that lose to kings, kings and queens, and all
 * face cards respectively. Following suit is not required, so the server sends
 * `validPlays` mainly to keep the devil out of the opening lead. Only `reset`
 * takes a `config`.
 */
export const karnoffelApi = {
  exec: (command: KarnoffelCommand, opts?: KarnoffelExecOptions) =>
    gameExec<KarnoffelResponse>('karnoffel', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
