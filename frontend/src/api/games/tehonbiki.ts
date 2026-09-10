import type { TehonbikiResponse } from '../../types/games/tehonbiki';
import { gameExec } from '../gameExec';

/** Calls the Tehonbiki game endpoint. */
export const tehonbikiApi = {
  exec: (
    command: 'reset' | 'bet' | 'next' | 'hint' | 'log',
    params?: { numbers?: number[]; betType?: string; bet?: number },
  ) => gameExec<TehonbikiResponse>('tehonbiki', { command, ...params }),
};
