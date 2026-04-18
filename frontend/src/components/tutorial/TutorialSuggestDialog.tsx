import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useBodyScrollLock } from '../../hooks/useBodyScrollLock';
import { btnPrimary, btnSecondary } from '../../styles/buttonStyles';
import { getFocusableElements } from '../../utils/dom';

/** Props for the TutorialSuggestDialog component. */
export interface TutorialSuggestDialogProps {
  /** Whether the dialog is visible. */
  open: boolean;
  /** Called when the user chooses to start the tutorial. */
  onStartTutorial: () => void;
  /** Called when the user skips the tutorial suggestion. */
  onSkip: () => void;
  /** Whether the "don't show again" checkbox is checked. */
  dontShowAgain: boolean;
  /** Called when the "don't show again" checkbox changes. */
  onDontShowAgainChange: (checked: boolean) => void;
}

/** Dialog shown on first visit to suggest starting the tutorial. */
export function TutorialSuggestDialog({
  open,
  onStartTutorial,
  onSkip,
  dontShowAgain,
  onDontShowAgainChange,
}: TutorialSuggestDialogProps) {
  const { t } = useTranslation('tutorial');
  const dialogRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<Element | null>(null);

  useBodyScrollLock(open);

  useEffect(() => {
    if (!open) return;
    triggerRef.current = document.activeElement;

    const dialog = dialogRef.current as HTMLElement;
    const focusable = getFocusableElements(dialog);
    if (focusable.length > 0) focusable[0].focus();

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;
      const elements = getFocusableElements(dialog);
      if (elements.length === 0) return;
      const first = elements[0];
      const last = elements[elements.length - 1];
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
  }, [open]);

  if (!open) return null;

  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: overlay backdrop dismisses dialog on click
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm"
      onClick={onSkip}
      role="presentation"
    >
      <div
        ref={dialogRef}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="suggest-dialog-title"
        aria-describedby="suggest-dialog-desc"
        className="glass-panel rounded-lg shadow-xl p-6 max-w-sm mx-4"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => {
          if (e.key === 'Escape') onSkip();
        }}
      >
        <h2 id="suggest-dialog-title" className="text-lg font-bold text-white mb-2">
          {t('firstVisit.title')}
        </h2>
        <p id="suggest-dialog-desc" className="text-ds-text-primary mb-4">
          {t('firstVisit.message')}
        </p>
        <label className="flex items-center gap-2 text-sm text-ds-text-muted mb-4 cursor-pointer">
          <input
            type="checkbox"
            checked={dontShowAgain}
            onChange={(e) => onDontShowAgainChange(e.target.checked)}
            className="rounded"
          />
          {t('firstVisit.dontShowAgain')}
        </label>
        <div className="flex justify-end gap-2">
          <button type="button" className={btnSecondary} onClick={onSkip}>
            {t('firstVisit.skip')}
          </button>
          <button type="button" className={btnPrimary} onClick={onStartTutorial}>
            {t('firstVisit.start')}
          </button>
        </div>
      </div>
    </div>
  );
}
