// API client for bigben. Split-file layout introduced by issue
// #4434; gameApi.ts re-exports this file, so existing imports keep working.

import type { BigBenMoveZone, BigBenResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Big Ben /bigben/exec endpoint. */
export const bigBenApi = createSolitaireMoveApi<
  BigBenResponse,
  BigBenMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('bigben');
