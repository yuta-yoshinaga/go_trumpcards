// API client for beleagueredcastle. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BeleagueredCastleMoveZone, BeleagueredCastleResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Beleaguered Castle /beleagueredcastle/exec endpoint. */
export const beleagueredCastleApi = createSolitaireMoveApi<
  BeleagueredCastleResponse,
  BeleagueredCastleMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('beleagueredcastle');
