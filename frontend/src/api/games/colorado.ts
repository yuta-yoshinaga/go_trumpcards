// API client for colorado. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ColoradoMoveZone, ColoradoResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Colorado /colorado/exec endpoint. */
export const coloradoApi = createSolitaireMoveApi<
  ColoradoResponse,
  ColoradoMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('colorado');
