import { useEffect, useRef } from 'react';
import { IGNORED_TAGS, isModalOpen } from './keyboardNavUtils';

/** Options for {@link useCardKeyboardNav}. */
export interface UseCardKeyboardNavOptions {
  cardCount: number;
  onToggle: (index: number) => void;
  onConfirm: () => void;
  onClear: () => void;
  enabled: boolean;
  onDirectPlay?: (index: number) => void;
  /** Returns whether a card may be played directly with a number key. */
  canDirectPlay?: (index: number) => boolean;
  /** Enables arrow-key hand navigation and Space-to-toggle for selection-only hands. */
  arrowSelection?: boolean;
  /** Reports the focused card index, or null when keyboard focus is cleared. */
  onFocusIndexChange?: (index: number | null) => void;
}

/** Hook that binds number keys to card selection and Enter/Escape to confirm/clear. */
export function useCardKeyboardNav({
  cardCount,
  onToggle,
  onConfirm,
  onClear,
  enabled,
  onDirectPlay,
  canDirectPlay,
  arrowSelection = false,
  onFocusIndexChange,
}: UseCardKeyboardNavOptions): void {
  const focusedIndexRef = useRef<number | null>(null);

  useEffect(() => {
    if (focusedIndexRef.current === null) return;
    if (cardCount === 0) {
      focusedIndexRef.current = null;
      onFocusIndexChange?.(null);
      return;
    }
    if (focusedIndexRef.current < cardCount) return;
    focusedIndexRef.current = Math.min(focusedIndexRef.current, cardCount - 1);
    onFocusIndexChange?.(focusedIndexRef.current);
  }, [cardCount, onFocusIndexChange]);

  useEffect(() => {
    if (!enabled) return;

    const handler = (e: KeyboardEvent) => {
      if (isModalOpen()) return;
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag && IGNORED_TAGS.has(tag)) return;

      if (e.key === 'Enter') {
        onConfirm();
        return;
      }
      if (e.key === 'Escape') {
        onClear();
        focusedIndexRef.current = null;
        onFocusIndexChange?.(null);
        return;
      }

      if (arrowSelection && (e.key === 'ArrowLeft' || e.key === 'ArrowRight')) {
        if (e.altKey || e.ctrlKey || e.metaKey || e.shiftKey) return;
        if (cardCount === 0) return;
        e.preventDefault();
        focusedIndexRef.current = Math.min(focusedIndexRef.current ?? 0, cardCount - 1);
        focusedIndexRef.current = Math.max(
          0,
          Math.min(cardCount - 1, focusedIndexRef.current + (e.key === 'ArrowRight' ? 1 : -1)),
        );
        onFocusIndexChange?.(focusedIndexRef.current);
        return;
      }

      if (arrowSelection && e.key === ' ') {
        if (tag === 'BUTTON' || cardCount === 0 || focusedIndexRef.current === null) return;
        e.preventDefault();
        onToggle(Math.min(focusedIndexRef.current, cardCount - 1));
        return;
      }

      const digit = Number.parseInt(e.key, 10);
      if (Number.isNaN(digit) || digit > 9) return;

      const index = digit === 0 ? 9 : digit - 1;
      if (index >= cardCount) return;

      if (onDirectPlay) {
        if (canDirectPlay && !canDirectPlay(index)) return;
        onDirectPlay(index);
      } else {
        onToggle(index);
      }
    };

    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [
    enabled,
    cardCount,
    onToggle,
    onConfirm,
    onClear,
    onDirectPlay,
    canDirectPlay,
    arrowSelection,
    onFocusIndexChange,
  ]);
}
