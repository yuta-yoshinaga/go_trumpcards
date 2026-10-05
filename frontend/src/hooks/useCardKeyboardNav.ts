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
  /** Enables arrow-key hand navigation and Enter-to-toggle for selection-only hands. */
  arrowSelection?: boolean;
  onFocusIndexChange?: (index: number) => void;
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
  const focusedIndexRef = useRef(0);
  const hasNavigatedRef = useRef(false);

  if (cardCount > 0 && focusedIndexRef.current >= cardCount) {
    focusedIndexRef.current = cardCount - 1;
  }

  useEffect(() => {
    if (!enabled) return;

    const handler = (e: KeyboardEvent) => {
      if (isModalOpen()) return;
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag && IGNORED_TAGS.has(tag)) return;

      if (e.key === 'Enter') {
        if (arrowSelection && hasNavigatedRef.current && cardCount > 0) {
          onToggle(focusedIndexRef.current);
          return;
        }
        onConfirm();
        return;
      }
      if (e.key === 'Escape') {
        onClear();
        return;
      }

      if (arrowSelection && (e.key === 'ArrowLeft' || e.key === 'ArrowRight')) {
        if (cardCount === 0) return;
        e.preventDefault();
        hasNavigatedRef.current = true;
        focusedIndexRef.current = Math.max(
          0,
          Math.min(cardCount - 1, focusedIndexRef.current + (e.key === 'ArrowRight' ? 1 : -1)),
        );
        onFocusIndexChange?.(focusedIndexRef.current);
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
