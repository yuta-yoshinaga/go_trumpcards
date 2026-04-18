import { useEffect, useId, useRef } from 'react';
import { btnDanger, btnSecondary } from '../styles/buttonStyles';
import { getFocusableElements } from '../utils/dom';

// Re-export for backward compatibility with existing imports
export { getFocusableElements } from '../utils/dom';

/** Props for the ConfirmDialog component. */
export interface ConfirmDialogProps {
  open: boolean;
  title: string;
  message: string;
  confirmLabel: string;
  cancelLabel: string;
  onConfirm: () => void;
  onCancel: () => void;
}

/** Renders a modal confirmation dialog with confirm and cancel buttons. */
export function ConfirmDialog(props: ConfirmDialogProps) {
  const dialogRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<Element | null>(null);
  const id = useId();
  const titleId = `${id}-title`;
  const descId = `${id}-desc`;
  const hasDescription = props.message.length > 0;

  useEffect(() => {
    if (!props.open) return;
    triggerRef.current = document.activeElement;

    // Safe assertion: useEffect runs after render, so ref is always attached when open=true
    const dialog = dialogRef.current as HTMLElement;
    const focusable = getFocusableElements(dialog);
    if (focusable.length === 0) return;

    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    first.focus();

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        props.onCancel();
        return;
      }
      if (e.key !== 'Tab') return;
      if (e.shiftKey) {
        if (document.activeElement === first) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      if (triggerRef.current instanceof HTMLElement) {
        triggerRef.current.focus();
      }
    };
  }, [props.open, props.onCancel]);

  if (!props.open) return null;

  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: overlay backdrop dismisses dialog on click
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={props.onCancel}
      role="presentation"
    >
      {/* biome-ignore lint/a11y/useKeyWithClickEvents: keyboard events handled at document level via useEffect */}
      <div
        ref={dialogRef}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={hasDescription ? descId : undefined}
        className="glass-panel rounded-lg shadow-xl p-6 max-w-sm mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id={titleId} className="text-lg font-bold text-ds-text-primary mb-2">
          {props.title}
        </h2>
        {hasDescription && (
          <p id={descId} className="text-ds-text-primary mb-4">
            {props.message}
          </p>
        )}
        <div className="flex justify-end gap-2">
          <button type="button" className={btnSecondary} onClick={props.onCancel}>
            {props.cancelLabel}
          </button>
          <button type="button" className={btnDanger} onClick={props.onConfirm}>
            {props.confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
