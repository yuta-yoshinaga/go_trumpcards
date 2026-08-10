import { useEffect } from 'react';
import { IGNORED_TAGS } from './keyboardNavUtils';

/** One keyboard shortcut bound to a game action. */
export interface ActionBinding {
  key: string;
  action: () => void;
  enabled?: boolean;
  /**
   * Which shared `kbd.action.*` label describes this shortcut. Supplying it
   * lets {@link components/ActionShortcutsPanel.ActionShortcutsPanel | ActionShortcutsPanel}
   * advertise the binding, so the list a player sees is generated from the same
   * array that binds the keys and cannot drift from it. Bindings without one are
   * bound but not advertised. See issue #4369.
   *
   * A label name rather than translated text deliberately: pages build these
   * arrays inside `useMemo`, and closing over a `t` function would add it to
   * every dependency list for no benefit.
   *
   * Typed `string` rather than a union of the known names: a union would force an
   * explicit `useMemo<ActionBinding[]>` annotation on all 56 pages, since these
   * array literals have no contextual type and would widen. The names are
   * validated instead by the kbd-label guard in scripts/check-design-tokens.mjs,
   * which fails the build if one has no `kbd.action.*` entry in common.json.
   */
  label?: string;
  /**
   * Interpolation values for the `kbd.action.*` label. Go Fish binds one key per
   * opponent and every row read "対象のプレイヤーを選ぶ", so the list could not
   * say which key meant which player (#4862). A label that names the target
   * needs the name, and the name is only known at binding time.
   */
  labelParams?: Record<string, string>;
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
