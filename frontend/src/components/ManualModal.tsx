import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import type { Components } from 'react-markdown';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { manualTexts } from '../constants/manualTexts';
import { btnSecondary } from '../styles/buttonStyles';
import { getFocusableElements } from '../utils/dom';
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
  const triggerRef = useRef<Element | null>(null);

  useEffect(() => {
    if (!open) return;
    triggerRef.current = document.activeElement;

    const dialog = dialogRef.current;
    if (!dialog) return;

    const focusable = getFocusableElements(dialog);
    if (focusable.length > 0) {
      focusable[0].focus();
    }

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
        return;
      }
      if (e.key !== 'Tab') return;
      const elems = getFocusableElements(dialog);
      if (elems.length === 0) return;
      const first = elems[0];
      const last = elems[elems.length - 1];
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

    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      if (triggerRef.current instanceof HTMLElement) {
        triggerRef.current.focus();
      }
    };
  }, [open, onClose]);

  if (!open) return null;

  const markdown = manualTexts[gamePath] ?? '';

  return (
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
        className="rounded-lg shadow-xl p-6 mx-4 max-w-2xl w-full max-h-[80vh] flex flex-col bg-gray-900 border border-gray-700"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex justify-end mb-2">
          <button type="button" className={btnSecondary} onClick={onClose} aria-label={t('manual.close')}>
            &times;
          </button>
        </div>
        <div className="overflow-y-auto flex-1 prose prose-invert prose-sm max-w-none">
          <Markdown remarkPlugins={remarkPlugins} components={markdownComponents}>
            {markdown}
          </Markdown>
        </div>
      </div>
    </div>
  );
}
