// API client for sultan. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SultanMoveZone, SultanResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Sultan of Turkey /sultan/exec endpoint. */
export const sultanApi = createSolitaireMoveApi<
  SultanResponse,
  SultanMoveZone,
  'reset' | 'draw' | 'redeal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('sultan');
