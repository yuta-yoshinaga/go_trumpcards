import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import type { actionLogApi } from '../api/gameApi';
import { SITE_NAME } from '../constants/site';
import { useActionLog } from './useActionLog';
import { useConfirmDialog } from './useConfirmDialog';

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

  const pageTitle = tc(`nav.${gameName}`);
  useEffect(() => {
    document.title = `${pageTitle} - ${SITE_NAME}`;
    return () => {
      document.title = SITE_NAME;
    };
  }, [pageTitle]);

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
