import { useEffect } from 'react';

interface UseCardKeyboardNavOptions {
  cardCount: number;
  onToggle: (index: number) => void;
  onConfirm: () => void;
  onClear: () => void;
  enabled: boolean;
  onDirectPlay?: (index: number) => void;
}

const IGNORED_TAGS = new Set(['INPUT', 'TEXTAREA', 'SELECT']);

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
      if (Number.isNaN(digit) || digit < 0 || digit > 9) return;

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
