import { useTranslation } from 'react-i18next';
import { useGameLeaveGuard } from './useGameLeaveGuard';

/**
 * Returns whether a round is in progress for the given game state.
 *
 * Accepts any state shape — most game responses expose `gameEndFlag`, but a
 * handful (solitaire and similar single-player games) only carry a `phase`
 * field. Callers can pass the state directly without per-page type
 * gymnastics; pages that need a stricter check (e.g., "round done but not
 * yet reset") should compute their own boolean and pass it instead.
 *
 * Returns false when state is null/undefined (no active round) or when
 * `gameEndFlag` is explicitly true. Otherwise returns true.
 */
export function isGameRoundActive(state: unknown): boolean {
  if (state == null || typeof state !== 'object') return false;
  if ('gameEndFlag' in state && (state as { gameEndFlag?: unknown }).gameEndFlag === true) {
    return false;
  }
  return true;
}

/**
 * Convenience wrapper around {@link useGameLeaveGuard} that resolves the
 * confirm message from the shared `button.confirmLeaveRoundMessage` i18n key,
 * removing the boilerplate from every game page.
 *
 * Pages should pass a single `isRoundInProgress` predicate computed from their
 * own phase state. When `true`, the browser-level `beforeunload` warning is
 * armed; SPA navigation is intentionally not intercepted (see the underlying
 * hook for rationale).
 *
 * Issue #1609 — applied across all long-form game pages so that accidental
 * tab close / reload during a round consistently warns the user, instead of
 * the prior BlackJack-only behavior.
 */
export function useGameRoundGuard(active: boolean) {
  const { t } = useTranslation('common');
  useGameLeaveGuard(active, t('button.confirmLeaveRoundMessage'));
}
