import { useCallback, useState } from 'react';
import type { CliLogEntry } from '../utils/cli/types';

const MAX_LOG_ENTRIES = 500;

/** Hook that manages CLI mode state and terminal log entries. */
export function useCliMode(gameName: string) {
  const storageKey = `cli-mode-${gameName}`;

  const [cliEnabled, setCliEnabled] = useState(() => {
    try {
      return localStorage.getItem(storageKey) === 'true';
    } catch {
      return false;
    }
  });

  const [logEntries, setLogEntries] = useState<CliLogEntry[]>([]);

  const toggleCli = useCallback(() => {
    setCliEnabled((prev) => {
      const next = !prev;
      try {
        localStorage.setItem(storageKey, String(next));
      } catch {
        /* ignore */
      }
      return next;
    });
  }, [storageKey]);

  const addEntry = useCallback((type: CliLogEntry['type'], text: string) => {
    setLogEntries((prev) => {
      const entry: CliLogEntry = { type, text, timestamp: Date.now() };
      const next = [...prev, entry];
      return next.length > MAX_LOG_ENTRIES ? next.slice(next.length - MAX_LOG_ENTRIES) : next;
    });
  }, []);

  const addInput = useCallback((text: string) => addEntry('input', text), [addEntry]);
  const addOutput = useCallback((text: string) => addEntry('output', text), [addEntry]);
  const addError = useCallback((text: string) => addEntry('error', text), [addEntry]);
  const clearLog = useCallback(() => setLogEntries([]), []);

  return { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog };
}
