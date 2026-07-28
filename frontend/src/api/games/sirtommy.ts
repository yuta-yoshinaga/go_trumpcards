// API client for sirtommy. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SirTommyMoveZone, SirTommyResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the SirTommy /sirtommy/exec endpoint. */
export const sirtommyApi = createSolitaireMoveApi<
  SirTommyResponse,
  SirTommyMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('sirtommy');
