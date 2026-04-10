import { useCallback, useEffect, useRef } from 'react';

/** Parameters for {@link useCardSwipeSelection}. */
export interface UseCardSwipeSelectionParams {
  /** Current selected card indices, used to decide swipe mode (select vs deselect). */
  selected: number[];
  /** Toggle a single card index in the existing selection. */
  toggle: (idx: number) => void;
  /** When false, pointer handlers are inert. */
  enabled?: boolean;
}

interface SwipeState {
  startIdx: number;
  mode: 'select' | 'deselect' | null;
  visited: Set<number>;
  moved: boolean;
}

/**
 * Enables swipe (drag) across card buttons to bulk select or deselect.
 *
 * Each card button must render with a `data-card-index="<index>"` attribute so
 * the hook can resolve which card is under the pointer via
 * `document.elementFromPoint`. Card buttons should also carry
 * `touch-action: none` (e.g. Tailwind's `touch-none`); otherwise mobile
 * browsers hijack the swipe as a scroll gesture and `pointermove` events
 * stop firing. The hook decides the swipe mode from the first card's current
 * selection state: if it's unselected, the swipe selects every card the
 * pointer traverses; if it's already selected, the swipe deselects them.
 * Single taps are left untouched so the existing click handler still toggles
 * one card at a time; the trailing click after an actual swipe is suppressed
 * to avoid double-toggling the card the pointer was released on.
 */
export function useCardSwipeSelection({ selected, toggle, enabled = true }: UseCardSwipeSelectionParams) {
  const stateRef = useRef<SwipeState | null>(null);
  const selectedRef = useRef(selected);
  const toggleRef = useRef(toggle);
  selectedRef.current = selected;
  toggleRef.current = toggle;

  const applyToIndex = useCallback((idx: number) => {
    const state = stateRef.current;
    if (!state || state.mode === null) return;
    if (state.visited.has(idx)) return;
    state.visited.add(idx);
    const has = selectedRef.current.includes(idx);
    if ((state.mode === 'select' && !has) || (state.mode === 'deselect' && has)) {
      toggleRef.current(idx);
    }
  }, []);

  useEffect(() => {
    if (!enabled) return;

    const findIndexAt = (clientX: number, clientY: number): number | null => {
      const el = document.elementFromPoint(clientX, clientY);
      if (!el) return null;
      const cardEl = (el as Element).closest('[data-card-index]');
      if (!cardEl) return null;
      const raw = cardEl.getAttribute('data-card-index');
      if (raw === null) return null;
      const n = Number(raw);
      return Number.isFinite(n) ? n : null;
    };

    const handleMove = (e: PointerEvent) => {
      const state = stateRef.current;
      if (!state) return;
      const idx = findIndexAt(e.clientX, e.clientY);
      if (idx === null) return;
      if (!state.moved) {
        if (idx === state.startIdx) return;
        // First move off the start card: fix the swipe mode based on the
        // start card's current selection state, then apply to both cards.
        const startSelected = selectedRef.current.includes(state.startIdx);
        state.mode = startSelected ? 'deselect' : 'select';
        state.moved = true;
        applyToIndex(state.startIdx);
      }
      applyToIndex(idx);
    };

    const suppressNextClick = () => {
      const suppress = (clickEvent: MouseEvent) => {
        // Only eat the click if it lands on a card; an unrelated click on a
        // nearby action button during the same task must still fire.
        const target = clickEvent.target as Element | null;
        if (target?.closest('[data-card-index]')) {
          clickEvent.stopPropagation();
          clickEvent.preventDefault();
        }
        window.removeEventListener('click', suppress, true);
      };
      window.addEventListener('click', suppress, true);
      // Fall through to the next task so the trusted click queued by the
      // browser's pointer sequence has already fired before we unregister;
      // this also cleans up when pointerup landed outside any card and no
      // click follows at all.
      setTimeout(() => window.removeEventListener('click', suppress, true), 0);
    };

    const handleUp = () => {
      const state = stateRef.current;
      if (!state) return;
      const wasSwipe = state.moved;
      stateRef.current = null;
      if (wasSwipe) suppressNextClick();
    };

    const handleDragStart = () => {
      // Native HTML5 drag took over (e.g. Daifugo's drag-to-play): abandon
      // the swipe so we don't double-dispatch when the drag ends.
      stateRef.current = null;
    };

    window.addEventListener('pointermove', handleMove);
    window.addEventListener('pointerup', handleUp);
    window.addEventListener('pointercancel', handleUp);
    window.addEventListener('dragstart', handleDragStart);
    return () => {
      window.removeEventListener('pointermove', handleMove);
      window.removeEventListener('pointerup', handleUp);
      window.removeEventListener('pointercancel', handleUp);
      window.removeEventListener('dragstart', handleDragStart);
    };
  }, [enabled, applyToIndex]);

  const onPointerDown = useCallback(
    (idx: number) => {
      if (!enabled) return;
      stateRef.current = {
        startIdx: idx,
        mode: null,
        visited: new Set<number>(),
        moved: false,
      };
    },
    [enabled],
  );

  return { onPointerDown } as const;
}
