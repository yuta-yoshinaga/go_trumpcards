import { useCallback, useEffect, useRef, useState } from 'react';

/** Options for {@link useTransientToast}. */
export interface UseTransientToastOptions {
  /**
   * When false, a `trigger` change does not show the toast (but still updates
   * the tracked previous value). Lets callers gate on state like "there is at
   * least one action to show". Defaults to true.
   */
  active?: boolean;
  /**
   * When true (default), the initial render never shows the toast — only later
   * `trigger` changes do. Set false to also treat the very first render as a
   * showable event (e.g. a component mounted with content already present).
   */
  skipInitial?: boolean;
}

/** Return value of {@link useTransientToast}. */
export interface TransientToast {
  /** Whether the toast should currently render. */
  visible: boolean;
  /** Dismiss the toast immediately (close button / programmatic). */
  dismiss: () => void;
}

/**
 * Drives the shared lifecycle of a transient banner toast: show whenever
 * `trigger` changes to a new value (skipping the initial render), auto-dismiss
 * after `durationMs`, allow early dismissal via Escape (unless a modal dialog
 * is open) or the returned `dismiss`, and restore focus to the element that had
 * it when the toast appeared. Extracted from the previously divergent
 * CpuActionToast / MetaAiToast implementations (issue #4313).
 */
export function useTransientToast(
  trigger: unknown,
  durationMs: number,
  { active = true, skipInitial = true }: UseTransientToastOptions = {},
): TransientToast {
  const [visible, setVisible] = useState(false);
  const prevRef = useRef(trigger);
  const isFirst = useRef(true);
  const timerRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const triggerElRef = useRef<Element | null>(null);

  useEffect(() => {
    const firstRun = isFirst.current;
    isFirst.current = false;
    if (firstRun && skipInitial) {
      prevRef.current = trigger;
      return;
    }
    if (trigger !== prevRef.current || (firstRun && !skipInitial)) {
      prevRef.current = trigger;
      if (active) {
        setVisible((prev) => {
          if (!prev) triggerElRef.current = document.activeElement;
          return true;
        });
        clearTimeout(timerRef.current);
        timerRef.current = setTimeout(() => setVisible(false), durationMs);
      }
    }
    return () => clearTimeout(timerRef.current);
  }, [trigger, durationMs, active, skipInitial]);

  useEffect(() => {
    if (!visible) {
      // Restore focus to the trigger element only if focus fell back to <body>
      // (i.e. nothing else claimed it) and the element is still in the DOM.
      if (
        document.activeElement === document.body &&
        triggerElRef.current instanceof HTMLElement &&
        document.contains(triggerElRef.current)
      ) {
        triggerElRef.current.focus();
      }
      triggerElRef.current = null;
      return;
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Escape') return;
      // Defer to an open modal dialog's own Escape handling.
      if (document.querySelector('[role="dialog"][aria-modal="true"]')) return;
      setVisible(false);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [visible]);

  const dismiss = useCallback(() => setVisible(false), []);
  return { visible, dismiss };
}
