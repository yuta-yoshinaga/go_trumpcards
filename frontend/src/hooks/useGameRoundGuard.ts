import { useTranslation } from 'react-i18next';
import { useGameLeaveGuard } from './useGameLeaveGuard';

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
