// API client for shamrocks. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ShamrocksResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Shamrocks /shamrocks/exec endpoint. */
export const shamrocksApi = createSolitaireMoveApi<
  ShamrocksResponse,
  number,
  'reset' | 'mf' | 'ff' | 'rd' | 'ac' | 'u' | 'undo_n' | 'giveup' | 'hint' | 'log'
>('shamrocks');
