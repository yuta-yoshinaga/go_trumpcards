// API client for curdsandwhey. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CurdsAndWheyResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands accepted by the Simple Simon /curdsandwhey/exec endpoint. */
export type CurdsAndWheyCommand = 'reset' | 'm' | 'g' | 'u' | 'undo_n' | 'hint' | 'log';

/** API client for the Simple Simon /curdsandwhey/exec endpoint. */
export const curdsandwheyApi = {
  exec: (command: CurdsAndWheyCommand, opts?: { fromCol?: number; cardIndex?: number; toCol?: number; n?: number }) =>
    gameExec<CurdsAndWheyResponse>('curdsandwhey', {
      command,
      fromCol: opts?.fromCol,
      cardIndex: opts?.cardIndex,
      toCol: opts?.toCol,
      n: opts?.n,
    }),
};
