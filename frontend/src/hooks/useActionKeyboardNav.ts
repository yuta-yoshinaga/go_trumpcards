import { useEffect } from 'react';
import { IGNORED_TAGS } from './keyboardNavUtils';

/** One keyboard shortcut bound to a game action. */
export interface ActionBinding {
  key: string;
  action: () => void;
  enabled?: boolean;
  /**
   * i18n key in the `common` namespace naming what this shortcut does, e.g.
   * `'kbd.action.fold'`. Supplying it lets
   * {@link components/ActionShortcutsPanel.ActionShortcutsPanel | ActionShortcutsPanel} advertise the
   * binding, so the list a player sees is generated from the same array that
   * binds the keys and cannot drift from it. Bindings without one are bound but
   * not advertised. See issue #4369.
   *
   * A key rather than translated text deliberately: pages build these arrays
   * inside `useMemo`, and closing over a `t` function would add it to every
   * dependency list for no benefit.
   */
  labelKey?: string;
}

/** Options for {@link useActionKeyboardNav}. */
export interface UseActionKeyboardNavOptions {
  bindings: ActionBinding[];
  enabled: boolean;
}

/** Hook that binds keyboard shortcuts to game actions. */
export function useActionKeyboardNav({ bindings, enabled }: UseActionKeyboardNavOptions): void {
  useEffect(() => {
    if (!enabled) return;

    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag && IGNORED_TAGS.has(tag)) return;

      const binding = bindings.find((b) => b.key === e.key);
      if (binding && binding.enabled !== false) {
        binding.action();
      }
    };

    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [enabled, bindings]);
}
