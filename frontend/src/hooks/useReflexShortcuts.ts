import { useEffect, useRef } from 'react';

/** Options for {@link useReflexShortcuts}. */
export interface UseReflexShortcutsOptions {
  /** Triggered when a step/flip key (Enter or S) is pressed. */
  onStep: () => void;
  /** Triggered when a slap key (Space) is pressed. */
  onSlap: () => void;
  /**
   * Master switch — when false the listener is detached entirely (CLI mode,
   * game over, etc.).
   */
  enabled: boolean;
  /**
   * Per-action gates so the keyboard shortcut mirrors the action button's
   * `disabled` state. Without these the hook would still fire `onStep` on
   * the CPU's turn even though the visible step button is greyed out.
   */
  stepEnabled?: boolean;
  /** Whether the slap shortcut is currently legal (e.g. centre pile is non-empty). */
  slapEnabled?: boolean;
}

/**
 * Global keyboard shortcuts for the real-time reflex games
 * (Slapjack, Egyptian Ratscrew). Wires Enter / S → step and Space → slap.
 *
 * Skips when the focus is in an editable element (or a focused
 * `<button>` / `<select>` — Space/Enter there must drive the native
 * activation, not the global game shortcut), when modifier keys are
 * held, and re-uses a ref so the listener does not re-bind on every
 * parent re-render (which happens roughly once per CPU tick).
 *
 * `stepEnabled` / `slapEnabled` default to `true` so existing callers
 * stay backward-compatible; pass them as the same expression that
 * disables the matching action button to keep keyboard and pointer UX
 * in lock-step.
 */
export function useReflexShortcuts({
  onStep,
  onSlap,
  enabled,
  stepEnabled = true,
  slapEnabled = true,
}: UseReflexShortcutsOptions): void {
  const handlersRef = useRef({ onStep, onSlap, stepEnabled, slapEnabled });
  handlersRef.current = { onStep, onSlap, stepEnabled, slapEnabled };

  useEffect(() => {
    if (!enabled) return;
    const handler = (event: KeyboardEvent): void => {
      // Modifiers reserve the keystroke for the browser / OS:
      //   - Shift+Space is "page-up", Shift+Enter is form submit-into-newline.
      //   - Ctrl/Alt/Meta combinations are user OS shortcuts.
      if (event.ctrlKey || event.altKey || event.metaKey || event.shiftKey) return;
      const target = event.target as HTMLElement | null;
      if (
        target &&
        (target.tagName === 'INPUT' ||
          target.tagName === 'TEXTAREA' ||
          target.tagName === 'SELECT' ||
          target.tagName === 'BUTTON' ||
          target.isContentEditable)
      ) {
        return;
      }

      const cur = handlersRef.current;
      if (event.key === ' ' || event.key === 'Spacebar') {
        if (!cur.slapEnabled) return;
        event.preventDefault();
        cur.onSlap();
        return;
      }
      if (event.key === 'Enter' || event.key === 's' || event.key === 'S') {
        if (!cur.stepEnabled) return;
        event.preventDefault();
        cur.onStep();
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [enabled]);
}
