// API client for guandan. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GuandanResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Guandan game settings. */
export interface GuandanConfigInput {
  cpuDifficulty?: number;
}

/** Commands the /guandan/exec endpoint accepts. */
export type GuandanCommand = 'reset' | 'play' | 'pass' | 'tribute' | 'next' | 'log';

/** Options carried alongside a Guandan command. */
export interface GuandanExecOptions {
  /** Hand indexes forming the combination. Required for `play`. */
  cardIndexes?: number[];
  /** The single card handed back. Required for `tribute`. */
  cardIndex?: number;
  config?: GuandanConfigInput;
}

/**
 * API client for the Guandan /guandan/exec endpoint.
 *
 * Guandan deals two full packs — 108 cards, 27 each — to four players in two
 * partnerships sitting opposite. Each hand is played at a **level**: cards of
 * that rank beat aces and lose only to the jokers, and the two hearts among
 * them are **wild**. Going out first and second climbs **four** levels, first
 * and third two, first and fourth one; there is no climb of three. Between
 * hands the losers pay **tribute**, unless a payer holds both red jokers.
 * Only `reset` takes a `config`.
 */
export const guandanApi = {
  exec: (command: GuandanCommand, opts?: GuandanExecOptions) =>
    gameExec<GuandanResponse>('guandan', {
      command,
      cardIndexes: opts?.cardIndexes,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
