import { useEffect, useId, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
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

function getFocusableElements(container: HTMLElement): HTMLElement[] {
  return Array.from(
    container.querySelectorAll<HTMLElement>(
      'a[href], button, input, select, textarea, [tabindex]:not([tabindex="-1"])',
    ),
  ).filter((el) => !el.hasAttribute('disabled'));
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

  useEffect(() => {
    triggerRef.current = document.activeElement;

    const dialog = dialogRef.current as HTMLElement;
    const focusable = getFocusableElements(dialog);
    if (focusable.length === 0) return;

    focusable[0].focus();
    const first = focusable[0];
    const last = focusable[focusable.length - 1];

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;
      if (e.shiftKey) {
        if (document.activeElement === first) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    dialog.addEventListener('keydown', handleKeyDown);
    return () => dialog.removeEventListener('keydown', handleKeyDown);
  }, []);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(textContent);
    setCopied(true);
  };

  const handleClose = () => {
    onClose();
    if (triggerRef.current instanceof HTMLElement) {
      triggerRef.current.focus();
    }
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
