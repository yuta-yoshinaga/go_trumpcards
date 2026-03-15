import { useEffect } from 'react';

interface ActionBinding {
  key: string;
  action: () => void;
  enabled?: boolean;
}

interface UseActionKeyboardNavOptions {
  bindings: ActionBinding[];
  enabled: boolean;
}

const IGNORED_TAGS = new Set(['INPUT', 'TEXTAREA', 'SELECT']);

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
