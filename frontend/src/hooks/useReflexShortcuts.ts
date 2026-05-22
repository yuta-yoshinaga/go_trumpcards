import { useEffect, useRef } from 'react';

/** Options for {@link useReflexShortcuts}. */
export interface UseReflexShortcutsOptions {
  /** Triggered when a step/flip key (Enter or S) is pressed. */
  onStep: () => void;
  /** Triggered when a slap key (Space) is pressed. */
  onSlap: () => void;
  /** When false the listener is detached (e.g. CLI mode, game over). */
  enabled: boolean;
}

/**
 * Global keyboard shortcuts for the real-time reflex games
 * (Slapjack, Egyptian Ratscrew). Wires Enter / S → step and Space → slap.
 *
 * Skips when the focus is in an editable element, when modifier keys are
 * held, and re-uses a ref so the listener does not re-bind on every parent
 * re-render (which happens roughly once per CPU tick).
 */
export function useReflexShortcuts({ onStep, onSlap, enabled }: UseReflexShortcutsOptions): void {
  const handlersRef = useRef({ onStep, onSlap });
  handlersRef.current = { onStep, onSlap };

  useEffect(() => {
    if (!enabled) return;
    const handler = (event: KeyboardEvent): void => {
      if (event.ctrlKey || event.altKey || event.metaKey) return;
      const target = event.target as HTMLElement | null;
      if (target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable)) return;

      if (event.key === ' ' || event.key === 'Spacebar') {
        event.preventDefault();
        handlersRef.current.onSlap();
        return;
      }
      if (event.key === 'Enter' || event.key === 's' || event.key === 'S') {
        event.preventDefault();
        handlersRef.current.onStep();
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [enabled]);
}
