import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import type { actionLogApi } from '../api/gameApi';
import { SITE_NAME } from '../constants/site';
import { useActionLog } from './useActionLog';
import { useConfirmDialog } from './useConfirmDialog';

/** Hook that provides common page setup: translations, action log, confirm dialog, and document title. */
export function useGamePageSetup(gameName: keyof typeof actionLogApi) {
  const { t } = useTranslation(gameName);
  const { t: tc } = useTranslation('common');
  const { actionLog, showActionLog, hideActionLog } = useActionLog(gameName);
  const { isOpen: confirmOpen, requestConfirm, confirm: confirmReset, cancel: cancelReset } = useConfirmDialog();

  const pageTitle = tc(`nav.${gameName}`);
  useEffect(() => {
    document.title = `${pageTitle} - ${SITE_NAME}`;
    return () => {
      document.title = SITE_NAME;
    };
  }, [pageTitle]);

  return { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset };
}
