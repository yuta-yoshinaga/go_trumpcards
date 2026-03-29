import { useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { btnSecondary } from '../styles/buttonStyles';
import { ManualModal } from './ManualModal';

/** Props for the ManualButton component. */
export interface ManualButtonProps {
  gamePath: string;
}

/** Renders a button that opens the game manual modal for the given game path. */
export function ManualButton({ gamePath }: ManualButtonProps) {
  const { t } = useTranslation('common');
  const [open, setOpen] = useState(false);
  const handleOpen = useCallback(() => setOpen(true), []);
  const handleClose = useCallback(() => setOpen(false), []);

  return (
    <>
      <button
        type="button"
        className={`${btnSecondary} text-xs`}
        onClick={handleOpen}
        aria-label={t('manual.button')}
        title={t('manual.button')}
      >
        {'\u{1F4D6}'}
      </button>
      <ManualModal open={open} onClose={handleClose} gamePath={gamePath} />
    </>
  );
}
