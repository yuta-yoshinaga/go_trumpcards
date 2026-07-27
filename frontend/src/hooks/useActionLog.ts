import { useCallback, useState } from 'react';
import { actionLogApi } from '../api/gameApi';
import type { ActionLogEntry } from '../types/card';
import { useIsMounted } from './useIsMounted';

/** Hook that provides action log fetching and display state for a game. */
export function useActionLog(game: keyof typeof actionLogApi) {
  const [actionLog, setActionLog] = useState<ActionLogEntry[] | null>(null);
  const isMounted = useIsMounted();

  const showActionLog = useCallback(async () => {
    const res = await actionLogApi[game]();
    // Closing the page mid-fetch must not write to it (#4447).
    if (!isMounted()) return;
    setActionLog(res.entries);
  }, [game, isMounted]);

  const hideActionLog = useCallback(() => {
    setActionLog(null);
  }, []);

  return { actionLog, showActionLog, hideActionLog };
}
