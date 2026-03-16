import { renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { usePhaseNames } from './usePhaseNames';

describe('usePhaseNames', () => {
  it('maps phase numbers to translated strings', () => {
    const phaseKeyMap: Readonly<Record<number, string>> = {
      0: 'flip1',
      1: 'flip2',
      2: 'result',
      3: 'gameEnd',
    };
    const { result } = renderHook(() => usePhaseNames('memory', phaseKeyMap));
    expect(result.current[0]).toBe('1枚目');
    expect(result.current[1]).toBe('2枚目');
    expect(result.current[2]).toBe('結果確認');
    expect(result.current[3]).toBe('ゲーム終了');
  });

  it('returns empty record for empty phaseKeyMap', () => {
    const { result } = renderHook(() => usePhaseNames('memory', {}));
    expect(result.current).toEqual({});
  });
});
