import type { ComponentPropsWithoutRef } from 'react';
import { useTranslation } from 'react-i18next';
import type { ActionBinding } from '../hooks/useActionKeyboardNav';
import { type KeyboardShortcut, KeyboardShortcutsPanel } from './KeyboardShortcutsPanel';

/** Props for the {@link ActionShortcutsPanel} component. */
export interface ActionShortcutsPanelProps extends Omit<ComponentPropsWithoutRef<'details'>, 'title'> {
  /**
   * The same bindings passed to
   * {@link hooks/useActionKeyboardNav.useActionKeyboardNav | useActionKeyboardNav}. Pages that call
   * the hook more than once (one array per phase) spread them all into one array
   * here.
   */
  bindings: ActionBinding[];
  /** Also list the fixed card-nav shortcuts, for pages that use both hooks. */
  includeCardNav?: boolean;
}

/**
 * Advertises the keyboard shortcuts a page bound through
 * {@link hooks/useActionKeyboardNav.useActionKeyboardNav | useActionKeyboardNav}.
 *
 * The list is derived from the bindings array itself, so an advertised key is by
 * construction a bound key. 94 pages implemented shortcuts without telling
 * anyone; the alternative of per-game `kbd.*` copy had already drifted before
 * this was written — `ja/blackjack.json` shipped `kbd.*` keys that no component
 * read. See issue #4369.
 *
 * A binding is listed only when it carries a `labelKey` and is not currently
 * disabled: `enabled: false` means pressing the key does nothing right now, so
 * listing it would advertise an action the player cannot take.
 */
export function ActionShortcutsPanel({ bindings, includeCardNav = false, ...rest }: ActionShortcutsPanelProps) {
  const { t } = useTranslation('common');
  const shortcuts: KeyboardShortcut[] = bindings
    .filter((b) => b.labelKey && b.enabled !== false)
    .map((b) => ({ keys: [b.key], description: t(b.labelKey as string) }));

  if (includeCardNav) {
    shortcuts.push(
      // `1`-`0` is rendered as a range by KeyboardShortcutsPanel.
      { keys: ['1', '0'], description: t('kbd.selectCard') },
      { keys: ['Enter'], description: t('kbd.confirm') },
      { keys: ['Esc'], description: t('kbd.clear') },
    );
  }

  // An empty <details> would open onto blank space, so render nothing instead.
  if (shortcuts.length === 0) return null;

  return <KeyboardShortcutsPanel title={t('kbd.title')} shortcuts={shortcuts} {...rest} />;
}
