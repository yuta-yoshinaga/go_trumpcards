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
  it('handles a successful command', async () => {
    const exec = vi.fn().mockResolvedValue(undefined);
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();
    const state: MockState = { score: 21 };

    const { result } = renderHook(() => useCliGame(exec, makeConfig(), state, { addInput, addOutput, addError }));

    await act(async () => {
      await result.current.handleCommand('hit');
    });

    expect(addInput).toHaveBeenCalledWith('hit');
    expect(exec).toHaveBeenCalledWith('hit');
    expect(addOutput).toHaveBeenCalledWith('score: 21');
    expect(addError).not.toHaveBeenCalled();
  });

  it('handles a command with arguments', async () => {
    const exec = vi.fn().mockResolvedValue(undefined);
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();
    const state: MockState = { score: 10 };

    const { result } = renderHook(() => useCliGame(exec, makeConfig(), state, { addInput, addOutput, addError }));

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

  it('handles exec error', async () => {
    const exec = vi.fn().mockRejectedValue(new Error('Network error'));
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();

    const { result } = renderHook(() => useCliGame(exec, makeConfig(), null, { addInput, addOutput, addError }));

    await act(async () => {
      await result.current.handleCommand('hit');
    });

    expect(addError).toHaveBeenCalled();
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

  it('does not format output when state is null after exec', async () => {
    const exec = vi.fn().mockResolvedValue(undefined);
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();

    const { result } = renderHook(() => useCliGame(exec, makeConfig(), null, { addInput, addOutput, addError }));

    await act(async () => {
      await result.current.handleCommand('hit');
    });

    // state is null, so formatResponse shouldn't produce output
    expect(addOutput).not.toHaveBeenCalled();
  });
});
