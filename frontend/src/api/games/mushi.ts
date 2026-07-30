// API client for mushi. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MushiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /mushi/exec endpoint accepts. */
export type MushiCommand = 'reset' | 'play' | 'select' | 'next' | 'hint' | 'log';

/**
 * API client for the Mushi /mushi/exec endpoint.
 *
 * `play` names a hand index; `select` names a FIELD index, used when two cards
 * of the same month are available or the lightning card needs a target.
 */
export const mushiApi = {
  exec: (command: MushiCommand, cardIndex?: number, fieldIndex?: number) =>
    gameExec<MushiResponse>('mushi', { command, cardIndex, fieldIndex }),
};
