import { useTranslation } from 'react-i18next';
import type { actionLogApi } from '../api/gameApi';
import { useActionLog } from './useActionLog';
import { useConfirmDialog } from './useConfirmDialog';
import { useDocumentTitle } from './useDocumentTitle';

/** Hook that provides common page setup: translations, action log, confirm dialogs, and document title. */
export function useGamePageSetup(gameName: keyof typeof actionLogApi) {
  const { t } = useTranslation(gameName);
  const { t: tc } = useTranslation('common');
  const { actionLog, showActionLog, hideActionLog } = useActionLog(gameName);
  const { isOpen: confirmOpen, requestConfirm, confirm: confirmReset, cancel: cancelReset } = useConfirmDialog();
  // Separate confirm instance for give-up so its dialog state never collides
  // with reset's. Give-up abandons an in-progress game and is irreversible,
  // so — like reset — it must be confirmed before firing (issue #2099).
  const {
    isOpen: giveUpConfirmOpen,
    requestConfirm: requestGiveUpConfirm,
    confirm: confirmGiveUp,
    cancel: cancelGiveUp,
  } = useConfirmDialog();

  // Title handling lives in useDocumentTitle so pages that are not games can
  // use it too -- before that split only game pages set a title (#5360).
  useDocumentTitle(tc(`nav.${gameName}`));

  return {
    t,
    tc,
    actionLog,
    showActionLog,
    hideActionLog,
    confirmOpen,
    requestConfirm,
    confirmReset,
    cancelReset,
    giveUpConfirmOpen,
    requestGiveUpConfirm,
    confirmGiveUp,
    cancelGiveUp,
  };
}
