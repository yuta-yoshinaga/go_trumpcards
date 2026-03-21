import { useEffect, useRef } from 'react';
import { btnDanger, btnSecondary } from '../styles/buttonStyles';

/** Returns all focusable elements within the given container. */
function getFocusableElements(container: HTMLElement): HTMLElement[] {
  return Array.from(
    container.querySelectorAll<HTMLElement>(
      'a[href], button, input, select, textarea, [tabindex]:not([tabindex="-1"])',
    ),
  ).filter((el) => !el.hasAttribute('disabled'));
}

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

    dialog.addEventListener('keydown', handleKeyDown);
    return () => {
      dialog.removeEventListener('keydown', handleKeyDown);
      if (triggerRef.current instanceof HTMLElement) {
        triggerRef.current.focus();
      }
    };
  }, [props.open]);

  if (!props.open) return null;

  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: overlay backdrop dismisses dialog on click
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={props.onCancel}
      role="presentation"
    >
      <div
        ref={dialogRef}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="confirm-dialog-title"
        className="glass-panel rounded-lg shadow-xl p-6 max-w-sm mx-4"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => {
          if (e.key === 'Escape') props.onCancel();
        }}
      >
        <h2 id="confirm-dialog-title" className="text-lg font-bold text-white mb-2">
          {props.title}
        </h2>
        <p className="text-gray-200 mb-4">{props.message}</p>
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
