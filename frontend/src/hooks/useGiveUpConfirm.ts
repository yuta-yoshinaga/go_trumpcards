import { useCallback } from 'react';

/**
 * Returns the button/key handler that routes a give-up action through the
 * give-up confirmation dialog. Collapses the `confirmGiveUpAction` `useCallback`
 * wrapper (`() => requestGiveUpConfirm(handleGiveUp)`) that was duplicated
 * across ~32 game pages (issue #4304).
 *
 * `giveUp` is the page's give-up action (however it sources it — an inline
 * `apiCall('giveup')` callback or one returned from a per-game hook);
 * `requestGiveUpConfirm` comes from {@link hooks/useGamePageSetup.useGamePageSetup | useGamePageSetup}. Give-up abandons an
 * in-progress game and is irreversible, so it is always confirmed before firing
 * (issue #2099).
 */
export function useGiveUpConfirm(giveUp: () => void, requestGiveUpConfirm: (action: () => void) => void): () => void {
  return useCallback(() => requestGiveUpConfirm(giveUp), [giveUp, requestGiveUpConfirm]);
}
