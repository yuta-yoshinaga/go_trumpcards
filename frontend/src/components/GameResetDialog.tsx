import { useTranslation } from 'react-i18next';
import { ConfirmDialog } from './ConfirmDialog';

/** Props for the GameResetDialog component. */
export interface GameResetDialogProps {
  confirmOpen: boolean;
  confirmReset: () => void;
  cancelReset: () => void;
}

/** Renders a reset confirmation dialog using common translation keys. */
export function GameResetDialog({ confirmOpen, confirmReset, cancelReset }: GameResetDialogProps) {
  const { t: tc } = useTranslation('common');
  return (
    <ConfirmDialog
      open={confirmOpen}
      title={tc('button.confirmReset')}
      message={tc('button.confirmResetMessage')}
      confirmLabel={tc('button.confirm')}
      cancelLabel={tc('button.cancel')}
      onConfirm={confirmReset}
      onCancel={cancelReset}
    />
  );
}
