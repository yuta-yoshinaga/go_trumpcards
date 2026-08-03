// API client for kille. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KilleResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Kille game settings. */
export interface KilleConfigInput {
  cpuDifficulty?: number;
  /** Stake each player antes per round (1-100). */
  stake?: number;
}

/** Commands the /kille/exec endpoint accepts. */
export type KilleCommand = 'reset' | 'exchange' | 'satisfied' | 'reenter' | 'nextround' | 'log';

/**
 * API client for the Kille /kille/exec endpoint.
 *
 * Kille is the Swedish Cuckoo game on its own 42-card pack. On your turn you
 * `exchange` with your left neighbour — who may not refuse — or declare yourself
 * `satisfied` and keep the card; the dealer swaps with the stock instead and so
 * cannot be challenged. After the showdown, a seat that went out may `reenter`
 * (three times at most, for one stake, then half the pot, then the whole pot),
 * and `nextround` deals again. Only `reset` takes a `config`.
 */
export const killeApi = {
  exec: (command: KilleCommand, opts?: { config?: KilleConfigInput }) =>
    gameExec<KilleResponse>('kille', {
      command,
      config: opts?.config,
    }),
};
