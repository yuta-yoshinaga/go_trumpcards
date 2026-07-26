import { useTranslation } from 'react-i18next';
import { btnSecondary } from '../styles/buttonStyles';
import type { ActionLogEntry } from '../types/card';
import { ActionLogPanel } from './ActionLogPanel';

/** Props for {@link ActionLogSection}. */
export interface ActionLogSectionProps {
  isEndPhase: boolean;
  actionLog: ActionLogEntry[] | null;
  showActionLog: () => void;
  hideActionLog: () => void;
}

/** Renders the action log view button and panel, shown at end phase. */
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
