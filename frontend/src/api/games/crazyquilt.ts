// API client for crazyquilt. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CrazyQuiltMoveZone, CrazyQuiltResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the CrazyQuilt /crazyquilt/exec endpoint. */
export const crazyquiltApi = createSolitaireMoveApi<
  CrazyQuiltResponse,
  CrazyQuiltMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('crazyquilt');
