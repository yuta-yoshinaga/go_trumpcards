// API client for simplesimon. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SimpleSimonResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands accepted by the Simple Simon /simplesimon/exec endpoint. */
export type SimpleSimonCommand = 'reset' | 'm' | 'g' | 'u' | 'undo_n' | 'hint' | 'log';

/** API client for the Simple Simon /simplesimon/exec endpoint. */
export const simplesimonApi = {
  exec: (command: SimpleSimonCommand, opts?: { fromCol?: number; cardIndex?: number; toCol?: number; n?: number }) =>
    gameExec<SimpleSimonResponse>('simplesimon', {
      command,
      fromCol: opts?.fromCol,
      cardIndex: opts?.cardIndex,
      toCol: opts?.toCol,
      n: opts?.n,
    }),
};
