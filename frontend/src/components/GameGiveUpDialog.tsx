import { useTranslation } from 'react-i18next';
import { ConfirmDialog } from './ConfirmDialog';

/** Props for the GameGiveUpDialog component. */
export interface GameGiveUpDialogProps {
  giveUpConfirmOpen: boolean;
  confirmGiveUp: () => void;
  cancelGiveUp: () => void;
}

/**
 * Renders a give-up confirmation dialog using common translation keys.
 * Mirrors {@link components/GameResetDialog.GameResetDialog | GameResetDialog} but with give-up-specific copy, since
 * give-up is an irreversible destructive action that deserves the same
 * "are you sure?" guard as reset (issue #2099).
 */
export function GameGiveUpDialog({ giveUpConfirmOpen, confirmGiveUp, cancelGiveUp }: GameGiveUpDialogProps) {
  const { t: tc } = useTranslation('common');
  return (
    <ConfirmDialog
      open={giveUpConfirmOpen}
      title={tc('button.confirmGiveUp')}
      message={tc('button.confirmGiveUpMessage')}
      confirmLabel={tc('button.confirm')}
      cancelLabel={tc('button.cancel')}
      onConfirm={confirmGiveUp}
      onCancel={cancelGiveUp}
    />
  );
}
