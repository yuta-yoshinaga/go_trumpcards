import { useTranslation } from 'react-i18next';
import { btnWarning } from '../styles/buttonStyles';

/** Props for the StalemateEscapeButton component. */
export interface StalemateEscapeButtonProps {
  undoToEscape: number;
  onEscape: (n: number) => void;
  disabled?: boolean;
}

/** Renders a button to batch-undo moves and escape a stalemate in solitaire games. */
export function StalemateEscapeButton({ undoToEscape, onEscape, disabled }: StalemateEscapeButtonProps) {
  const { t } = useTranslation('common');

  if (undoToEscape <= 0) return null;

  return (
    <button
      type="button"
      className={`${btnWarning} animate-pulse`}
      onClick={() => onEscape(undoToEscape)}
      disabled={disabled}
      data-testid="stalemate-escape-button"
    >
      {t('stalemateEscape', { count: undoToEscape })}
    </button>
  );
}
