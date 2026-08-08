import { useEffect, useId, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import type { ActionLogEntry } from '../types/card';
import { cardLabel } from '../utils/cardUtils';
import { getFocusableElements } from '../utils/dom';

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
  const triggerRef = useRef<Element | null>(null);
  const titleId = useId();

  const textContent = entries.length === 0 ? t('actionLog.empty') : entries.map((e) => formatEntry(e, t)).join('\n');

  useEffect(() => {
    if (!copied) return;
    const timer = setTimeout(() => setCopied(false), 2000);
    return () => clearTimeout(timer);
  }, [copied]);

  // Keep the latest onClose without re-running the effect (and re-stealing
  // focus) when a parent passes a new inline callback each render.
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

  // This panel is a landmark `role="region"`, not a dialog — no `aria-modal`,
  // the game behind it stays live, and it exists to be read alongside the
  // board. So it deliberately does NOT trap Tab: cycling focus inside a
  // non-modal region is a WCAG 2.1.2 keyboard trap, and the only way out was
  // to find the close button by sight. Moving focus in on open is still right
  // (the user opened it deliberately); Escape and focus restore give them the
  // way back. See issue #5183.
  useEffect(() => {
    triggerRef.current = document.activeElement;

    const dialog = dialogRef.current;
    if (dialog) getFocusableElements(dialog)[0]?.focus();

    // Listen on document, not on the panel: once focus legitimately leaves the
    // region, a panel-level listener would never see the key again.
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onCloseRef.current();
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      // Restore here rather than in the close handler so any unmount path
      // (game reset, route change) returns focus, not just the close button.
      if (triggerRef.current instanceof HTMLElement) triggerRef.current.focus();
    };
  }, []);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(textContent);
    setCopied(true);
  };

  // Focus restore lives in the effect cleanup, which covers this path too —
  // the page closes the panel by clearing its state, so onClose unmounts it.
  const handleClose = () => {
    onClose();
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
          <button type="button" className={btnPrimary} onClick={handleClose}>
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
