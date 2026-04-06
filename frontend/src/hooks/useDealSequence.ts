import { useCallback, useRef, useState } from 'react';

/** Deal sequence state. */
export type DealState = 'idle' | 'dealing' | 'dealt';

/** Options for the deal sequence hook. */
export interface DealSequenceOptions {
  /** Number of cards to deal. */
  count: number;
  /** Stagger delay between cards in seconds. Default: 0.12. */
  stagger?: number;
}

/** Return value of useDealSequence. */
export interface DealSequenceResult {
  /** Current deal state. */
  state: DealState;
  /** Get the deal delay for a card at the given index. Returns 0 when not dealing. */
  getDelay: (index: number) => number;
  /** Start the deal sequence. Transitions from idle to dealing. */
  startDeal: () => void;
  /** Reset to idle state. */
  reset: () => void;
}

/**
 * Orchestrates staggered card dealing with state management.
 * Games use `state === 'dealing'` to disable buttons during deal animation.
 */
export function useDealSequence({ count, stagger = 0.12 }: DealSequenceOptions): DealSequenceResult {
  const [state, setState] = useState<DealState>('idle');
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const startDeal = useCallback(() => {
    if (count <= 0) return;
    setState('dealing');
    if (timerRef.current) clearTimeout(timerRef.current);
    const totalDuration = count * stagger * 1000 + 300; // stagger + settle time (matches dealSpring ~300ms)
    timerRef.current = setTimeout(() => {
      setState('dealt');
      timerRef.current = null;
    }, totalDuration);
  }, [count, stagger]);

  const reset = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
    setState('idle');
  }, []);

  const getDelay = useCallback(
    (index: number) => {
      if (state !== 'dealing') return 0;
      return index * stagger;
    },
    [state, stagger],
  );

  return { state, getDelay, startDeal, reset };
}
