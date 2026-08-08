import { useEffect, useRef } from 'react';
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
  const triggerRef = useRef<HTMLButtonElement>(null);
  const wasOpen = useRef(false);

  // Focus restore has to live here, not in ActionLogPanel. Opening the panel
  // sets `actionLog`, which unmounts the trigger button below — so by the time
  // the panel records "what had focus", the button is already gone and the
  // panel's own restore aims at a detached node. Closing remounts the trigger
  // as a *new* element, which only this component can reach. Measured: without
  // this, focus after close is on neither the old nor the new trigger.
  // See issue #5183.
  useEffect(() => {
    const isOpen = actionLog !== null;
    if (wasOpen.current && !isOpen) triggerRef.current?.focus();
    wasOpen.current = isOpen;
  }, [actionLog]);

  return (
    <>
      {isEndPhase && !actionLog && (
        <div className="text-center my-2">
          <button type="button" ref={triggerRef} className={btnSecondary} onClick={showActionLog}>
            {tc('actionLog.view')}
          </button>
        </div>
      )}
      {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
    </>
  );
}
