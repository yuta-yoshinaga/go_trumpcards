import { useEffect, useId, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useFocusTrap } from '../hooks/useFocusTrap';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import type { ActionLogEntry } from '../types/card';
import { cardLabel } from '../utils/cardUtils';

/** Props for {@link ActionLogPanel}. */
export interface ActionLogPanelProps {
  entries: ActionLogEntry[];
  onClose: () => void;
}

function formatEntry(entry: ActionLogEntry, t: (key: string, opts?: Record<string, unknown>) => string): string {
  const player = entry.playerIdx < 0 ? t('actionLog.system') : t('actionLog.player', { idx: entry.playerIdx });
  let line = `T${entry.turnNumber} [${player}] ${entry.actionType}: ${entry.detail}`;
  if (entry.cards && entry.cards.length > 0) {
    line += ` [${entry.cards.map(cardLabel).join(', ')}]`;
  }
  return line;
}

/** Renders a panel displaying game action log entries with copy and download. */
export function ActionLogPanel({ entries, onClose }: ActionLogPanelProps) {
  const { t } = useTranslation('common');
  const [copied, setCopied] = useState(false);
  const dialogRef = useRef<HTMLElement>(null);
  const titleId = useId();

  const textContent = entries.length === 0 ? t('actionLog.empty') : entries.map((e) => formatEntry(e, t)).join('\n');

  useEffect(() => {
    if (!copied) return;
    const timer = setTimeout(() => setCopied(false), 2000);
    return () => clearTimeout(timer);
  }, [copied]);

  // Landmark `role="region"`, not a dialog: no `aria-modal`, the game behind it
  // stays live, and it exists to be read alongside the board. So Tab is NOT
  // cycled — that would be a WCAG 2.1.2 keyboard trap with no way out but
  // finding the close button by sight (issue #5183). Everything else the hook
  // provides is still wanted: focus in on open, Escape to leave, and restore on
  // any unmount path (game reset, route change), not just the close button.
  useFocusTrap(dialogRef, true, onClose, { trap: false });

  const handleCopy = async () => {
    await navigator.clipboard.writeText(textContent);
    setCopied(true);
  };

  const handleDownload = () => {
    const blob = new Blob([textContent], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'action_log.txt';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <section
      ref={dialogRef}
      aria-labelledby={titleId}
      className="glass-panel rounded-lg p-4 my-2 max-h-[60vh] flex flex-col"
    >
      <div className="flex items-center justify-between mb-2">
        <h3 id={titleId} className="text-ds-text-primary font-bold text-sm">
          {t('actionLog.title')}
        </h3>
        <div className="flex gap-2">
          <button type="button" className={btnSecondary} onClick={handleCopy}>
            {copied ? t('actionLog.copied') : t('actionLog.copy')}
          </button>
          <button type="button" className={btnSecondary} onClick={handleDownload}>
            {t('actionLog.download')}
          </button>
          <button type="button" className={btnPrimary} onClick={onClose}>
            {t('actionLog.close')}
          </button>
        </div>
      </div>
      <div data-testid="copy-announcer" aria-live="polite" aria-atomic="true" className="sr-only">
        {copied ? t('actionLog.copied') : ''}
      </div>
      <pre className="flex-1 overflow-y-auto text-ds-text-primary text-xs whitespace-pre-wrap font-mono bg-black/40 rounded p-2">
        {textContent}
      </pre>
    </section>
  );
}
