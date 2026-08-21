// API client for perseverance. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PerseveranceMoveZone, PerseveranceResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Perseverance /perseverance/exec endpoint. */
export const perseveranceApi = createSolitaireMoveApi<
  PerseveranceResponse,
  PerseveranceMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n' | 'redeal'
>('perseverance');
