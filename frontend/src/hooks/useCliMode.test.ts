import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { useCliMode } from './useCliMode';

describe('useCliMode', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('defaults to CLI disabled', () => {
    const { result } = renderHook(() => useCliMode('blackjack'));
    expect(result.current.cliEnabled).toBe(false);
  });

  it('toggles CLI mode', () => {
    const { result } = renderHook(() => useCliMode('blackjack'));
    act(() => result.current.toggleCli());
    expect(result.current.cliEnabled).toBe(true);
    act(() => result.current.toggleCli());
    expect(result.current.cliEnabled).toBe(false);
  });

  it('persists CLI mode to localStorage', () => {
    const { result } = renderHook(() => useCliMode('blackjack'));
    act(() => result.current.toggleCli());
    expect(localStorage.getItem('cli-mode-blackjack')).toBe('true');
  });

  it('restores CLI mode from localStorage', () => {
    localStorage.setItem('cli-mode-poker', 'true');
    const { result } = renderHook(() => useCliMode('poker'));
    expect(result.current.cliEnabled).toBe(true);
  });

  it('starts with empty log', () => {
    const { result } = renderHook(() => useCliMode('blackjack'));
    expect(result.current.logEntries).toEqual([]);
  });

  it('adds input entry', () => {
    const { result } = renderHook(() => useCliMode('blackjack'));
    act(() => result.current.addInput('hit'));
    expect(result.current.logEntries).toHaveLength(1);
    expect(result.current.logEntries[0]?.type).toBe('input');
    expect(result.current.logEntries[0].text).toBe('hit');
  });

  it('adds output entry', () => {
    const { result } = renderHook(() => useCliMode('blackjack'));
    act(() => result.current.addOutput('score: 21'));
    expect(result.current.logEntries).toHaveLength(1);
    expect(result.current.logEntries[0]?.type).toBe('output');
  });

  it('adds error entry', () => {
    const { result } = renderHook(() => useCliMode('blackjack'));
    act(() => result.current.addError('Unknown command'));
    expect(result.current.logEntries).toHaveLength(1);
    expect(result.current.logEntries[0]?.type).toBe('error');
  });

  it('clears log', () => {
    const { result } = renderHook(() => useCliMode('blackjack'));
    act(() => result.current.addInput('hit'));
    act(() => result.current.addOutput('ok'));
    act(() => result.current.clearLog());
    expect(result.current.logEntries).toEqual([]);
  });

  it('caps log at 500 entries', () => {
    const { result } = renderHook(() => useCliMode('blackjack'));
    act(() => {
      for (let i = 0; i < 510; i++) {
        result.current.addOutput(`line ${i}`);
      }
    });
    expect(result.current.logEntries.length).toBeLessThanOrEqual(500);
  });
});
