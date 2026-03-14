import { useTranslation } from 'react-i18next';
import { btnSecondary } from '../styles/buttonStyles';
import type { ActionLogEntry } from '../types/card';
import { ActionLogPanel } from './ActionLogPanel';

interface ActionLogSectionProps {
  isEndPhase: boolean;
  actionLog: ActionLogEntry[] | null;
  showActionLog: () => void;
  hideActionLog: () => void;
}

export function ActionLogSection({ isEndPhase, actionLog, showActionLog, hideActionLog }: ActionLogSectionProps) {
  const { t: tc } = useTranslation('common');
  return (
    <>
      {isEndPhase && !actionLog && (
        <div className="text-center my-2">
          <button type="button" className={btnSecondary} onClick={showActionLog}>
            {tc('actionLog.view')}
          </button>
        </div>
      )}
      {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
    </>
  );
}
