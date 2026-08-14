// API client for diplomat. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { DiplomatMoveZone, DiplomatResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Diplomat /diplomat/exec endpoint. */
export const diplomatApi = createSolitaireMoveApi<
  DiplomatResponse,
  DiplomatMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('diplomat');
