import { useTranslation } from 'react-i18next';
import type { actionLogApi } from '../api/gameApi';
import { useActionLog } from './useActionLog';
import { useConfirmDialog } from './useConfirmDialog';

/** Hook that provides common page setup: translations, action log, and confirm dialog. */
export function useGamePageSetup(gameName: keyof typeof actionLogApi) {
  const { t } = useTranslation(gameName);
  const { t: tc } = useTranslation('common');
  const { actionLog, showActionLog, hideActionLog } = useActionLog(gameName);
  const { isOpen: confirmOpen, requestConfirm, confirm: confirmReset, cancel: cancelReset } = useConfirmDialog();

  return { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset };
}
