import { useEffect } from 'react';
import { IGNORED_TAGS } from './keyboardNavUtils';

/** One keyboard shortcut bound to a game action. */
export interface ActionBinding {
  key: string;
  action: () => void;
  enabled?: boolean;
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
