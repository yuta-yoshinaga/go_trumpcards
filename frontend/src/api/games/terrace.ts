// API client for terrace. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TerraceMoveZone, TerraceResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Terrace /terrace/exec endpoint. */
export const terraceApi = createSolitaireMoveApi<
  TerraceResponse,
  TerraceMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('terrace');
