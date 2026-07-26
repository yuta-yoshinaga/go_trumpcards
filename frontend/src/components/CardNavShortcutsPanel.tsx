import type { ComponentPropsWithoutRef } from 'react';
import { useTranslation } from 'react-i18next';
import { type KeyboardShortcut, KeyboardShortcutsPanel } from './KeyboardShortcutsPanel';

/** Props for the {@link CardNavShortcutsPanel} component. */
export interface CardNavShortcutsPanelProps extends Omit<ComponentPropsWithoutRef<'details'>, 'title'> {
  /**
   * Set on pages that pass `onDirectPlay` to
   * {@link hooks/useCardKeyboardNav.useCardKeyboardNav | useCardKeyboardNav}. There, a number key
   * plays the card outright instead of toggling its selection, so Enter and
   * Escape do nothing and are omitted rather than advertised falsely.
   */
  directPlay?: boolean;
  /** Extra rows for keys the page binds itself, appended after the card-nav set. */
  extra?: KeyboardShortcut[];
}

/**
 * Advertises the keyboard shortcuts that
 * {@link hooks/useCardKeyboardNav.useCardKeyboardNav | useCardKeyboardNav} binds on every page that
 * uses it.
 *
 * That hook's bindings are fixed — `1`–`9` and `0` address the first ten cards,
 * `Enter` confirms, `Escape` clears — so the copy lives in `common.json` once
 * rather than being restated in each game's namespace. 38 pages implemented
 * these shortcuts without telling anyone they existed; per-game copy would have
 * meant 38 chances for the advertised keys to drift from the bound ones. See
 * issue #4369.
 */
export function CardNavShortcutsPanel({ directPlay = false, extra = [], ...rest }: CardNavShortcutsPanelProps) {
  const { t } = useTranslation('common');
  const shortcuts: KeyboardShortcut[] = directPlay
    ? // `1`–`0` is rendered as a range by KeyboardShortcutsPanel.
      [{ keys: ['1', '0'], description: t('kbd.playCard') }]
    : [
        { keys: ['1', '0'], description: t('kbd.selectCard') },
        { keys: ['Enter'], description: t('kbd.confirm') },
        { keys: ['Esc'], description: t('kbd.clear') },
      ];
  return <KeyboardShortcutsPanel title={t('kbd.title')} shortcuts={[...shortcuts, ...extra]} {...rest} />;
}
