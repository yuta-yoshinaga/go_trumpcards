import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useGameSound } from './useGameSound';

describe('useGameSound', () => {
  let mockPlay: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    localStorage.clear();
    mockPlay = vi.fn().mockResolvedValue(undefined);
    vi.spyOn(globalThis, 'Audio').mockImplementation(
      function () {
        return {
          play: mockPlay,
          volume: 0,
        } as unknown as HTMLAudioElement;
      } as unknown as typeof Audio,
    );
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('defaults to unmuted', () => {
    const { result } = renderHook(() => useGameSound());
    expect(result.current.muted).toBe(false);
  });

  it('reads muted state from localStorage', () => {
    localStorage.setItem('trumpcards-sound-muted', 'true');
    const { result } = renderHook(() => useGameSound());
    expect(result.current.muted).toBe(true);
  });

  it('toggleMute switches muted state', () => {
    const { result } = renderHook(() => useGameSound());
    expect(result.current.muted).toBe(false);
    act(() => result.current.toggleMute());
    expect(result.current.muted).toBe(true);
    act(() => result.current.toggleMute());
    expect(result.current.muted).toBe(false);
  });

  it('persists muted state to localStorage', () => {
    const { result } = renderHook(() => useGameSound());
    act(() => result.current.toggleMute());
    expect(localStorage.getItem('trumpcards-sound-muted')).toBe('true');
  });

  it('playCardDeal plays when unmuted', () => {
    const { result } = renderHook(() => useGameSound());
    act(() => result.current.playCardDeal());
    expect(mockPlay).toHaveBeenCalled();
  });

  it('playCardFlip plays when unmuted', () => {
    const { result } = renderHook(() => useGameSound());
    act(() => result.current.playCardFlip());
    expect(mockPlay).toHaveBeenCalled();
  });

  it('playSelect plays when unmuted', () => {
    const { result } = renderHook(() => useGameSound());
    act(() => result.current.playSelect());
    expect(mockPlay).toHaveBeenCalled();
  });

  it('playWin plays when unmuted', () => {
    const { result } = renderHook(() => useGameSound());
    act(() => result.current.playWin());
    expect(mockPlay).toHaveBeenCalled();
  });

  it('does not play when muted', () => {
    localStorage.setItem('trumpcards-sound-muted', 'true');
    const { result } = renderHook(() => useGameSound());
    act(() => result.current.playCardDeal());
    expect(mockPlay).not.toHaveBeenCalled();
  });

  it('handles Audio constructor failure gracefully', () => {
    vi.spyOn(globalThis, 'Audio').mockImplementation(
      function () {
        throw new Error('Audio not supported');
      } as unknown as typeof Audio,
    );
    const { result } = renderHook(() => useGameSound());
    act(() => result.current.playCardDeal());
    // No error thrown
  });

  it('handles play rejection gracefully', () => {
    mockPlay.mockRejectedValue(new Error('Autoplay blocked'));
    const { result } = renderHook(() => useGameSound());
    act(() => result.current.playCardDeal());
    // No error thrown
  });

  it('handles localStorage error on read gracefully', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('localStorage blocked');
    });
    const { result } = renderHook(() => useGameSound());
    expect(result.current.muted).toBe(false);
  });

  it('handles localStorage error on write gracefully', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('localStorage blocked');
    });
    const { result } = renderHook(() => useGameSound());
    act(() => result.current.toggleMute());
    // No error thrown
  });
});
