import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { CliLogEntry } from '../../utils/cli/types';

const MAX_HISTORY = 50;

/** Props for the CliTerminal component. */
export interface CliTerminalProps {
  /** Log entries to display. */
  logEntries: CliLogEntry[];
  /** Callback when the user submits a command. */
  onCommand: (command: string) => void;
  /** Whether the input is disabled (e.g., while loading). */
  disabled: boolean;
}

const ENTRY_STYLES: Record<CliLogEntry['type'], string> = {
  input: 'text-ds-success',
  output: 'text-ds-text-primary whitespace-pre-wrap',
  error: 'text-ds-error',
};

/** Renders a pseudo-terminal UI with log display and command input. */
export function CliTerminal({ logEntries, onCommand, disabled }: CliTerminalProps) {
  const { t } = useTranslation('common');
  const [inputValue, setInputValue] = useState('');
  const historyRef = useRef<string[]>([]);
  const historyIdxRef = useRef(-1);
  const scrollRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Auto-scroll to bottom when new entries arrive
  // biome-ignore lint/correctness/useExhaustiveDependencies: scroll must trigger on entry count change
  useEffect(() => {
    scrollRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logEntries.length]);

  // Focus input on mount
  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  const onCommandRef = useRef(onCommand);
  onCommandRef.current = onCommand;

  const handleSubmit = useCallback(() => {
    const trimmed = inputRef.current?.value.trim() ?? '';
    if (!trimmed) return;
    onCommandRef.current(trimmed);
    // Add to history
    const history = historyRef.current;
    if (history.length >= MAX_HISTORY) history.shift();
    history.push(trimmed);
    historyIdxRef.current = -1;
    setInputValue('');
  }, []);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        handleSubmit();
        return;
      }
      const history = historyRef.current;
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        if (history.length === 0) return;
        const nextIdx = historyIdxRef.current === -1 ? history.length - 1 : Math.max(0, historyIdxRef.current - 1);
        historyIdxRef.current = nextIdx;
        setInputValue(history[nextIdx]);
        return;
      }
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        if (history.length === 0) return;
        if (historyIdxRef.current === -1) return;
        const nextIdx = historyIdxRef.current + 1;
        if (nextIdx >= history.length) {
          historyIdxRef.current = -1;
          setInputValue('');
        } else {
          historyIdxRef.current = nextIdx;
          setInputValue(history[nextIdx]);
        }
      }
    },
    [handleSubmit],
  );

  // SR announcement policy (issue #1843): the visible log captures the full
  // session and stays focusable so SR users can step into it on demand. The
  // separate live region below carries only the *latest* important entry so
  // screen readers don't recite every line every turn. Inputs are user-typed
  // and would be echoed back as noise, so they are skipped here.
  const latestAnnouncement = useMemo(() => {
    for (let i = logEntries.length - 1; i >= 0; i--) {
      const e = logEntries[i];
      if (e.type === 'error') return e.text;
      if (e.type === 'output') {
        // Games typically end output with a summary line (e.g. "Your turn",
        // "BUST", "+50 chips"). Announce just that last line to keep the
        // utterance short and actionable, not a wall of state.
        const lines = e.text.trim().split('\n');
        return lines[lines.length - 1] ?? '';
      }
    }
    return '';
  }, [logEntries]);

  return (
    <div className="flex flex-col flex-1 bg-black rounded-lg border border-ds-border overflow-hidden font-mono text-sm min-h-[300px]">
      {/* Log display area: focusable for SR users to step into; aria-live
          is explicitly "off" to override role="log"'s implicit "polite"
          default — auto-announcement of every entry would flood the
          screen reader during normal play (issue #1843). */}
      <div
        className="flex-1 overflow-y-auto p-3 space-y-1"
        role="log"
        aria-live="off"
        aria-label={t('cli.logLabel')}
        // biome-ignore lint/a11y/noNoninteractiveTabindex: scrollable log region must be reachable by keyboard so SR users (and keyboard-only users) can step through history; required by WCAG 2.1.1.
        tabIndex={0}
      >
        {logEntries.map((entry) => (
          <div key={entry.id} className={ENTRY_STYLES[entry.type]}>
            {entry.type === 'input' ? `> ${entry.text}` : entry.text}
          </div>
        ))}
        <div ref={scrollRef} />
      </div>
      {/* Off-screen polite live region announcing only the latest important
          entry — errors announce in full, outputs announce their last line.
          Inputs are skipped to avoid echoing the user's own typing. */}
      <div className="sr-only" aria-live="polite" aria-atomic="true">
        {latestAnnouncement}
      </div>
      {/* Command input */}
      <div className="flex items-center border-t border-ds-border px-3 py-2 bg-black">
        <span className="text-ds-success mr-2" aria-hidden="true">
          &gt;
        </span>
        <input
          ref={inputRef}
          type="text"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={disabled}
          placeholder={t('cli.prompt')}
          className="flex-1 bg-transparent text-ds-text-primary outline-none placeholder:text-ds-text-muted min-h-[44px] py-2"
          aria-label={t('cli.prompt')}
          autoComplete="off"
          spellCheck={false}
        />
      </div>
    </div>
  );
}
