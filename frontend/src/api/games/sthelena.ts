// API client for sthelena. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { StHelenaMoveZone, StHelenaResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the StHelena Solitaire /sthelena/exec endpoint. */
export const stHelenaApi = createSolitaireMoveApi<
  StHelenaResponse,
  StHelenaMoveZone,
  'reset' | 'move' | 'redeal' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('sthelena');
