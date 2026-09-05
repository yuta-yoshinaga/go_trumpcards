// API client for citadel. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CitadelMoveZone, CitadelResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

export type { CitadelMoveZone, CitadelResponse };

/** API client for the Beleaguered Castle /citadel/exec endpoint. */
export const citadelApi = createSolitaireMoveApi<
  CitadelResponse,
  CitadelMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('citadel');
