import { useTranslation } from 'react-i18next';
import { btnPrimary, btnSecondary } from '../../styles/buttonStyles';
import { Modal } from '../common/Modal';

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
  return (
    <Modal
      open={open}
      onClose={onSkip}
      role="alertdialog"
      ariaLabelledBy="suggest-dialog-title"
      ariaDescribedBy="suggest-dialog-desc"
      panelClassName="glass-panel rounded-lg shadow-xl p-6 max-w-sm mx-4"
      backdropClassName="items-center justify-center bg-black/80 backdrop-blur-sm"
    >
      <h2 id="suggest-dialog-title" className="text-lg font-bold text-ds-text-primary mb-2">
        {t('firstVisit.title')}
      </h2>
      <p id="suggest-dialog-desc" className="text-ds-text-primary mb-4">
        {t('firstVisit.message')}
      </p>
      <label className="flex items-center gap-2 text-sm text-ds-text-muted mb-4 cursor-pointer min-h-[44px]">
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
    </Modal>
  );
}
