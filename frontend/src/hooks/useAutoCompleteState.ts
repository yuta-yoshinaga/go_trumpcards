import { useCallback, useEffect, useRef, useState } from 'react';

/** Default duration (ms) for the auto-complete animation window. */
const AUTO_COMPLETE_DURATION_MS = 3000;

/**
 * Manages the isAutoCompleting flag with a self-clearing timeout.
 * Shared by Klondike, FreeCell, and Spider solitaire hooks.
 */
export function useAutoCompleteState(duration = AUTO_COMPLETE_DURATION_MS) {
  const [isAutoCompleting, setIsAutoCompleting] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const startAutoComplete = useCallback(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
    setIsAutoCompleting(true);
    timerRef.current = setTimeout(() => setIsAutoCompleting(false), duration);
  }, [duration]);

  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  return { isAutoCompleting, startAutoComplete };
}
