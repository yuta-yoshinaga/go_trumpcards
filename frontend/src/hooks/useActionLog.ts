import { useCallback, useState } from 'react';
import { actionLogApi } from '../api/gameApi';
import type { ActionLogEntry } from '../types/card';

/** Hook that provides action log fetching and display state for a game. */
export function useActionLog(game: keyof typeof actionLogApi) {
  const [actionLog, setActionLog] = useState<ActionLogEntry[] | null>(null);

  const showActionLog = useCallback(async () => {
    const res = await actionLogApi[game]();
    setActionLog(res.entries);
  }, [game]);

  const hideActionLog = useCallback(() => {
    setActionLog(null);
  }, []);

  return { actionLog, showActionLog, hideActionLog };
}
