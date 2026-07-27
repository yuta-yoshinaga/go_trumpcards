// API client for trash. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TrashResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Command verbs accepted by the Trash /trash/exec endpoint. */
export type TrashCommand = 'reset' | 'draw' | 'place' | 'cpu' | 'log';

/** API client for the Trash /trash/exec endpoint. */
export const trashApi = {
  exec: (command: TrashCommand, position?: number) => gameExec<TrashResponse>('trash', { command, position }),
};
