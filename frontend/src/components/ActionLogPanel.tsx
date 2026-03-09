import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import type { ActionLogEntry } from '../types/card';
import { cardLabel } from '../utils/cardUtils';

interface ActionLogPanelProps {
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

export function ActionLogPanel({ entries, onClose }: ActionLogPanelProps) {
  const { t } = useTranslation('common');
  const [copied, setCopied] = useState(false);

  const textContent = entries.length === 0 ? t('actionLog.empty') : entries.map((e) => formatEntry(e, t)).join('\n');

  useEffect(() => {
    if (!copied) return;
    const timer = setTimeout(() => setCopied(false), 2000);
    return () => clearTimeout(timer);
  }, [copied]);

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
    <div className="bg-black/60 rounded-lg p-4 my-2 max-h-[60vh] flex flex-col">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-white font-bold text-sm">{t('actionLog.title')}</h3>
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
      <pre className="flex-1 overflow-y-auto text-[#ccc] text-xs whitespace-pre-wrap font-mono bg-black/40 rounded p-2">
        {textContent}
      </pre>
    </div>
  );
}
