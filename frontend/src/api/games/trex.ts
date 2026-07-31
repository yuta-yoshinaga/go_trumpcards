// API client for trex. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TrexResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /trex/exec endpoint accepts. */
export type TrexCommand = 'reset' | 'choose' | 'play' | 'pass' | 'next' | 'hint' | 'log';

/**
 * API client for the Trex /trex/exec endpoint.
 *
 * `contract` and `cardIndex` are separate parameters so a contract number can
 * never be read as a hand index. Contract 0 is the king of hearts — a real
 * value, not an omission.
 */
export const trexApi = {
  exec: (command: TrexCommand, contract?: number, cardIndex?: number) =>
    gameExec<TrexResponse>('trex', { command, contract, cardIndex }),
};
