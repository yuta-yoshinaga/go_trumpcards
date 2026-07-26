import { useCallback, useEffect, useRef } from 'react';
import type { CliGameConfig } from '../utils/cli/types';

/** Log callbacks provided by useCliMode. */
export interface CliLogCallbacks {
  addInput: (text: string) => void;
  addOutput: (text: string) => void;
  addError: (text: string) => void;
  clearLog?: () => void;
}

/** Hook that wires CLI command input to a game API exec function. */
export function useCliGame<TState, TArgs extends unknown[]>(
  exec: (...args: TArgs) => Promise<void>,
  config: CliGameConfig<TState, TArgs>,
  state: TState | null,
  callbacks: CliLogCallbacks,
) {
  const configRef = useRef(config);
  configRef.current = config;
  const callbacksRef = useRef(callbacks);
  callbacksRef.current = callbacks;
  const execRef = useRef(exec);
  execRef.current = exec;
  const pendingCommandRef = useRef(false);

  // Format and output state when it changes after a CLI command
  useEffect(() => {
    if (pendingCommandRef.current && state !== null) {
      pendingCommandRef.current = false;
      callbacksRef.current.addOutput(configRef.current.formatResponse(state));
    }
  }, [state]);

  const handleCommand = useCallback(async (input: string) => {
    const { addInput, addOutput, addError, clearLog } = callbacksRef.current;
    const cfg = configRef.current;
    const cmd = input.trim().toLowerCase();

    // Built-in commands
    if (cmd === 'help' || cmd === '?') {
      addInput(input);
      addOutput(cfg.helpText.join('\n'));
      return;
    }
    if (cmd === 'clear') {
      clearLog?.();
      return;
    }

    addInput(input);

    // Parse command
    const parsed = cfg.parseCommand(input);
    if ('error' in parsed) {
      addError(parsed.error);
      return;
    }

    // Execute API call — state will be formatted in useEffect after re-render
    try {
      pendingCommandRef.current = true;
      await execRef.current(...parsed.args);
    } catch (e) {
      pendingCommandRef.current = false;
      addError(e instanceof Error ? e.message : 'Error executing command');
    }
  }, []);

  return { handleCommand };
}
