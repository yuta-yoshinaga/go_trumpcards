import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { CliGameConfig } from '../utils/cli/types';
import { useCliGame } from './useCliGame';

type MockState = { score: number };
type MockArgs = [string, number?];

function makeConfig(overrides?: Partial<CliGameConfig<MockState, MockArgs>>): CliGameConfig<MockState, MockArgs> {
  return {
    gameName: 'test',
    parseCommand: (input) => {
      if (input === 'hit') return { args: ['hit'] };
      if (input.startsWith('bet ')) return { args: ['bet', Number(input.split(' ')[1])] };
      return { error: `Unknown: ${input}` };
    },
    formatResponse: (state) => `score: ${state.score}`,
    helpText: ['hit - draw card', 'bet <n> - place bet'],
    ...overrides,
  };
}

describe('useCliGame', () => {
  it('calls exec with parsed args on valid command', async () => {
    const exec = vi.fn().mockResolvedValue(undefined);
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();

    const { result } = renderHook(() =>
      useCliGame(exec, makeConfig(), { score: 21 }, { addInput, addOutput, addError }),
    );

    await act(async () => {
      await result.current.handleCommand('hit');
    });

    expect(addInput).toHaveBeenCalledWith('hit');
    expect(exec).toHaveBeenCalledWith('hit');
  });

  it('formats output via useEffect when state changes after command', async () => {
    const exec = vi.fn().mockResolvedValue(undefined);
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();
    const { result, rerender } = renderHook(
      ({ s }: { s: MockState | null }) => useCliGame(exec, makeConfig(), s, { addInput, addOutput, addError }),
      { initialProps: { s: null as MockState | null } },
    );

    await act(async () => {
      await result.current.handleCommand('hit');
    });

    // State hasn't changed yet — no output
    expect(addOutput).not.toHaveBeenCalled();

    // Simulate state update (as would happen after exec updates useGameApi state)
    rerender({ s: { score: 21 } });

    expect(addOutput).toHaveBeenCalledWith('score: 21');
  });

  it('handles a command with arguments', async () => {
    const exec = vi.fn().mockResolvedValue(undefined);
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();

    const { result } = renderHook(() =>
      useCliGame(exec, makeConfig(), { score: 10 }, { addInput, addOutput, addError }),
    );

    await act(async () => {
      await result.current.handleCommand('bet 100');
    });

    expect(exec).toHaveBeenCalledWith('bet', 100);
  });

  it('handles parse error', async () => {
    const exec = vi.fn();
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();

    const { result } = renderHook(() => useCliGame(exec, makeConfig(), null, { addInput, addOutput, addError }));

    await act(async () => {
      await result.current.handleCommand('xyz');
    });

    expect(addInput).toHaveBeenCalledWith('xyz');
    expect(exec).not.toHaveBeenCalled();
    expect(addError).toHaveBeenCalledWith('Unknown: xyz');
  });

  it('handles exec error with message', async () => {
    const exec = vi.fn().mockRejectedValue(new Error('Network error'));
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();

    const { result } = renderHook(() => useCliGame(exec, makeConfig(), null, { addInput, addOutput, addError }));

    await act(async () => {
      await result.current.handleCommand('hit');
    });

    expect(addError).toHaveBeenCalledWith('Network error');
  });

  it('handles exec error without message', async () => {
    const exec = vi.fn().mockRejectedValue('unknown');
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();

    const { result } = renderHook(() => useCliGame(exec, makeConfig(), null, { addInput, addOutput, addError }));

    await act(async () => {
      await result.current.handleCommand('hit');
    });

    expect(addError).toHaveBeenCalledWith('Error executing command');
  });

  it('shows help text', async () => {
    const exec = vi.fn();
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();

    const { result } = renderHook(() => useCliGame(exec, makeConfig(), null, { addInput, addOutput, addError }));

    await act(async () => {
      await result.current.handleCommand('help');
    });

    expect(addInput).toHaveBeenCalledWith('help');
    expect(addOutput).toHaveBeenCalledWith('hit - draw card\nbet <n> - place bet');
    expect(exec).not.toHaveBeenCalled();
  });

  it('handles ? as help alias', async () => {
    const exec = vi.fn();
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();

    const { result } = renderHook(() => useCliGame(exec, makeConfig(), null, { addInput, addOutput, addError }));

    await act(async () => {
      await result.current.handleCommand('?');
    });

    expect(addOutput).toHaveBeenCalled();
    expect(exec).not.toHaveBeenCalled();
  });

  it('handles clear command', async () => {
    const exec = vi.fn();
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();
    const clearLog = vi.fn();

    const { result } = renderHook(() =>
      useCliGame(exec, makeConfig(), null, { addInput, addOutput, addError, clearLog }),
    );

    await act(async () => {
      await result.current.handleCommand('clear');
    });

    expect(clearLog).toHaveBeenCalled();
    expect(exec).not.toHaveBeenCalled();
  });
});
