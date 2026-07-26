import { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { useTranslation } from 'react-i18next';
import type { Components } from 'react-markdown';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { isCliModeEnabled, loadCuiManualText } from '../constants/cuiManualTexts';
import { loadManualText } from '../constants/manualTexts';
import { useBodyScrollLock } from '../hooks/useBodyScrollLock';
import { useFocusTrap } from '../hooks/useFocusTrap';
import { btnSecondary } from '../styles/buttonStyles';
import { MermaidBlock } from './MermaidBlock';

/** Props for the ManualModal component. */
export interface ManualModalProps {
  open: boolean;
  onClose: () => void;
  gamePath: string;
}

/** Static remark plugins array — defined outside component to avoid re-creating on every render. */
const remarkPlugins = [remarkGfm];

/** Static markdown component overrides — defined outside component to avoid re-rendering the entire tree on every render. */
const markdownComponents: Components = {
  pre({ children }) {
    // Unwrap <pre> when the child is a mermaid diagram rendered by the code override
    const child = Array.isArray(children) ? children[0] : children;
    if (child && typeof child === 'object' && 'type' in child && child.type === MermaidBlock) {
      return <>{children}</>;
    }
    return <pre>{children}</pre>;
  },
  code({ className, children }) {
    if (className === 'language-mermaid') {
      return <MermaidBlock code={String(children).trim()} />;
    }
    return <code className={className}>{children}</code>;
  },
};

/** Renders a scrollable modal displaying the game manual as rendered Markdown. */
export function ManualModal({ open, onClose, gamePath }: ManualModalProps) {
  const { t } = useTranslation('common');
  const dialogRef = useRef<HTMLDivElement>(null);
  // Tagged with what it is a manual OF, not just the text. Holding a bare
  // string would render the previous game's manual for a frame when `gamePath`
  // changes while the modal is open, and would serve the web manual after the
  // reader toggles CLI mode. Comparing the tag makes a mismatch render the
  // loading state instead, and still reuses the text when reopening the same
  // manual (Vite caches the chunk, so that resolves in a microtask).
  const [loaded, setLoaded] = useState<{ key: string; text: string } | null>(null);

  useBodyScrollLock(open);
  useFocusTrap(dialogRef, open, onClose);

  const cliMode = isCliModeEnabled(gamePath);
  const manualKey = `${gamePath}|${cliMode ? 'cui' : 'web'}`;

  // Each manual is its own chunk now, so it arrives a tick after the modal
  // opens rather than being already in memory. `cancelled` covers the reader
  // closing the modal or switching games before the fetch lands.
  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    const load = cliMode ? loadCuiManualText : loadManualText;
    load(gamePath).then((text) => {
      if (!cancelled) setLoaded({ key: `${gamePath}|${cliMode ? 'cui' : 'web'}`, text });
    });
    return () => {
      cancelled = true;
    };
  }, [open, gamePath, cliMode]);

  if (!open) return null;

  const ready = loaded?.key === manualKey;
  const markdown = ready ? loaded.text : '';

  return createPortal(
    // biome-ignore lint/a11y/noStaticElementInteractions: overlay backdrop dismisses modal on click
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={onClose}
      role="presentation"
    >
      {/* biome-ignore lint/a11y/useKeyWithClickEvents: keyboard events handled at document level via useEffect */}
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-label={t('manual.ariaLabel')}
        className="rounded-lg shadow-xl p-6 mx-4 max-w-4xl w-full h-[calc(100vh-4rem)] supports-[height:100dvh]:h-[calc(100dvh-4rem)] overflow-hidden flex flex-col bg-ds-surface border border-ds-border-subtle"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex justify-end mb-2 flex-shrink-0">
          <button type="button" className={btnSecondary} onClick={onClose} aria-label={t('manual.close')}>
            &times;
          </button>
        </div>
        <div className="overflow-y-auto flex-1 min-h-0 prose prose-invert max-w-none">
          {ready ? (
            <Markdown remarkPlugins={remarkPlugins} components={markdownComponents}>
              {markdown}
            </Markdown>
          ) : (
            // The chunk is small and same-origin, so this is usually one frame —
            // but it is a fetch, and on a cold cache over a slow link it is not.
            // Mirrors RouteSuspenseFallback's role="status" / aria-busy contract.
            <p role="status" aria-busy="true" className="text-ds-text-muted">
              {t('skeleton.loading')}
            </p>
          )}
        </div>
      </div>
    </div>,
    document.body,
  );
}
