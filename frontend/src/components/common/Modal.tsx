import { type ReactNode, useRef } from 'react';
import { createPortal } from 'react-dom';
import { useBodyScrollLock } from '../../hooks/useBodyScrollLock';
import { useFocusTrap } from '../../hooks/useFocusTrap';

/** Props for the shared {@link Modal} primitive. */
export interface ModalProps {
  /** Whether the modal is shown. */
  open: boolean;
  /** Called on Escape, backdrop click (unless disabled), or programmatic close. */
  onClose: () => void;
  /** Dialog content. */
  children: ReactNode;
  /** ARIA role — use 'alertdialog' for confirmations, 'dialog' otherwise. */
  role?: 'dialog' | 'alertdialog';
  /** Accessible name when there is no visible titled element to reference. */
  ariaLabel?: string;
  /** id of the element labelling the dialog (preferred over ariaLabel). */
  ariaLabelledBy?: string;
  /** id of the element describing the dialog. */
  ariaDescribedBy?: string;
  /** Classes for the inner dialog panel. */
  panelClassName?: string;
  /** Whether a backdrop click closes the modal (default true). */
  dismissOnBackdrop?: boolean;
}

/**
 * Shared modal-dialog primitive (issue #4312): a single implementation of the
 * a11y behavior that had diverged across eight hand-rolled dialogs. It portals
 * to `document.body` (so it escapes `overflow-hidden` game panels), renders a
 * dismiss-on-click backdrop scrim, locks body scroll via {@link useBodyScrollLock},
 * and traps focus / closes on Escape / restores focus via {@link useFocusTrap}.
 * Centered-dialog layout; bespoke overlays (manual, tutorial spotlight) reuse
 * the `useFocusTrap` hook directly instead.
 */
export function Modal({
  open,
  onClose,
  children,
  role = 'dialog',
  ariaLabel,
  ariaLabelledBy,
  ariaDescribedBy,
  panelClassName = '',
  dismissOnBackdrop = true,
}: ModalProps) {
  const dialogRef = useRef<HTMLDivElement>(null);
  useBodyScrollLock(open);
  useFocusTrap(dialogRef, open, onClose);

  if (!open) return null;

  return createPortal(
    // biome-ignore lint/a11y/noStaticElementInteractions: overlay backdrop dismisses the dialog on click
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      role="presentation"
      onClick={dismissOnBackdrop ? onClose : undefined}
    >
      {/* biome-ignore lint/a11y/useKeyWithClickEvents: keyboard handling lives in useFocusTrap (document-level) */}
      {/* biome-ignore lint/a11y/noStaticElementInteractions: this is the dialog panel (role dialog/alertdialog); onClick only stops backdrop propagation */}
      {/* biome-ignore lint/a11y/useAriaPropsSupportedByRole: role is constrained to dialog|alertdialog, both of which support aria-modal (biome can't narrow the dynamic prop) */}
      <div
        ref={dialogRef}
        role={role}
        aria-modal="true"
        aria-label={ariaLabel}
        aria-labelledby={ariaLabelledBy}
        aria-describedby={ariaDescribedBy}
        className={panelClassName}
        onClick={(e) => e.stopPropagation()}
      >
        {children}
      </div>
    </div>,
    document.body,
  );
}
