import { useEffect } from 'react';
import { IGNORED_TAGS } from './keyboardNavUtils';

/** Options for {@link useCardKeyboardNav}. */
export interface UseCardKeyboardNavOptions {
  cardCount: number;
  onToggle: (index: number) => void;
  onConfirm: () => void;
  onClear: () => void;
  enabled: boolean;
  onDirectPlay?: (index: number) => void;
}

/** Hook that binds number keys to card selection and Enter/Escape to confirm/clear. */
export function useCardKeyboardNav({
  cardCount,
  onToggle,
  onConfirm,
  onClear,
  enabled,
  onDirectPlay,
}: UseCardKeyboardNavOptions): void {
  useEffect(() => {
    if (!enabled) return;

    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag && IGNORED_TAGS.has(tag)) return;

      if (e.key === 'Enter') {
        onConfirm();
        return;
      }
      if (e.key === 'Escape') {
        onClear();
        return;
      }

      const digit = Number.parseInt(e.key, 10);
      if (Number.isNaN(digit) || digit > 9) return;

      const index = digit === 0 ? 9 : digit - 1;
      if (index >= cardCount) return;

      if (onDirectPlay) {
        onDirectPlay(index);
      } else {
        onToggle(index);
      }
    };

    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [enabled, cardCount, onToggle, onConfirm, onClear, onDirectPlay]);
}
