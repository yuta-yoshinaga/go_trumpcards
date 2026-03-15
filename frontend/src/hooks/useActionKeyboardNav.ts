import { useEffect } from 'react';
import { IGNORED_TAGS } from './keyboardNavUtils';

interface ActionBinding {
  key: string;
  action: () => void;
  enabled?: boolean;
}

interface UseActionKeyboardNavOptions {
  bindings: ActionBinding[];
  enabled: boolean;
}

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
