import { useCallback, useRef } from 'react';
import type { CliGameConfig } from '../utils/cli/types';

/** Log callbacks provided by useCliMode. */
interface CliLogCallbacks {
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
  const stateRef = useRef(state);
  stateRef.current = state;
  const configRef = useRef(config);
  configRef.current = config;
  const callbacksRef = useRef(callbacks);
  callbacksRef.current = callbacks;
  const execRef = useRef(exec);
  execRef.current = exec;

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

    // Execute API call
    try {
      await execRef.current(...parsed.args);
      // Format response from updated state
      const currentState = stateRef.current;
      if (currentState !== null) {
        addOutput(cfg.formatResponse(currentState));
      }
    } catch {
      addError('Error executing command');
    }
  }, []);

  return { handleCommand };
}
