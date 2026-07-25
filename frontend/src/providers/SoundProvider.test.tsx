import { act, render, renderHook, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { SOUND_ENABLED, SoundProvider, useSound } from './SoundProvider';

const { mockPlay, mockRate, mockVolume, mockHowlerCtx } = vi.hoisted(() => ({
  mockPlay: vi.fn().mockReturnValue(1),
  mockRate: vi.fn(),
  mockVolume: vi.fn(),
  mockHowlerCtx: { state: 'running' },
}));

vi.mock('howler', () => ({
  Howl: class MockHowl {
    play = mockPlay;
    rate = mockRate;
    volume = mockVolume;
  },
  Howler: { ctx: mockHowlerCtx },
}));

function wrapper({ children }: { children: React.ReactNode }) {
  return <SoundProvider>{children}</SoundProvider>;
}

describe('SoundProvider', () => {
  beforeEach(() => {
    localStorage.clear();
    mockPlay.mockClear().mockReturnValue(1);
    mockRate.mockClear();
    mockVolume.mockClear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('context access', () => {
    it('provides playSound, muted, and toggleMute via useSound', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      expect(result.current.playSound).toBeTypeOf('function');
      expect(result.current.muted).toBe(false);
      expect(result.current.toggleMute).toBeTypeOf('function');
    });

    it('throws when useSound is called outside SoundProvider', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      expect(() => renderHook(() => useSound())).toThrow('useSound must be used within a SoundProvider');
      consoleSpy.mockRestore();
    });
  });

  describe('mute state', () => {
    it('defaults to unmuted', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      expect(result.current.muted).toBe(false);
    });

    it('reads muted state from localStorage', () => {
      localStorage.setItem('trumpcards-sound-muted', 'true');
      const { result } = renderHook(() => useSound(), { wrapper });
      expect(result.current.muted).toBe(true);
    });

    it('toggles muted state', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      expect(result.current.muted).toBe(false);
      act(() => result.current.toggleMute());
      expect(result.current.muted).toBe(true);
      act(() => result.current.toggleMute());
      expect(result.current.muted).toBe(false);
    });

    it('persists muted state to localStorage', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.toggleMute());
      expect(localStorage.getItem('trumpcards-sound-muted')).toBe('true');
    });

    it('handles localStorage read error gracefully', () => {
      vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
        throw new Error('localStorage blocked');
      });
      const { result } = renderHook(() => useSound(), { wrapper });
      expect(result.current.muted).toBe(false);
    });

    it('handles localStorage write error gracefully', () => {
      vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
        throw new Error('localStorage blocked');
      });
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.toggleMute());
      expect(result.current.muted).toBe(true);
    });
  });

  describe('playSound', () => {
    it('plays sound when unmuted', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('cardDeal'));
      expect(mockPlay).toHaveBeenCalled();
    });

    it('does not play when muted', () => {
      localStorage.setItem('trumpcards-sound-muted', 'true');
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('cardDeal'));
      expect(mockPlay).not.toHaveBeenCalled();
    });

    it('applies pitch variation when specified', () => {
      mockPlay.mockReturnValue(42);
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('cardDeal', { pitchVariation: 0.05 }));
      expect(mockPlay).toHaveBeenCalled();
      expect(mockRate).toHaveBeenCalled();
      const rate = mockRate.mock.calls[0][0] as number;
      expect(rate).toBeGreaterThanOrEqual(0.95);
      expect(rate).toBeLessThanOrEqual(1.05);
    });

    it('applies volume override when specified', () => {
      mockPlay.mockReturnValue(42);
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('cardDeal', { volume: 0.8 }));
      expect(mockVolume).toHaveBeenCalledWith(0.8, 42);
    });

    it('applies default volume tier for ambient sounds', () => {
      mockPlay.mockReturnValue(42);
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('cardDeal'));
      // Ambient tier: 0.6x base (0.3) = 0.18
      expect(mockVolume).toHaveBeenCalledWith(expect.closeTo(0.18, 1), 42);
    });

    it('applies default volume tier for action sounds', () => {
      mockPlay.mockReturnValue(42);
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('cardFlip'));
      // Action tier: 1.0x base (0.3) = 0.30
      expect(mockVolume).toHaveBeenCalledWith(expect.closeTo(0.3, 1), 42);
    });

    it('applies default volume tier for event sounds', () => {
      mockPlay.mockReturnValue(42);
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('winFanfare'));
      // Event tier: 1.4x base (0.3) = 0.42
      expect(mockVolume).toHaveBeenCalledWith(expect.closeTo(0.42, 1), 42);
    });
  });

  describe('policy layer: throttle', () => {
    beforeEach(() => vi.useFakeTimers());
    afterEach(() => vi.useRealTimers());

    it('suppresses a second same-sound play within its min interval', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('cardPlace'));
      act(() => result.current.playSound('cardPlace'));
      expect(mockPlay).toHaveBeenCalledTimes(1);
    });

    it('plays again after the min interval elapses', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('cardPlace'));
      act(() => vi.advanceTimersByTime(100));
      act(() => result.current.playSound('cardPlace'));
      expect(mockPlay).toHaveBeenCalledTimes(2);
    });

    it('throttles winFanfare with the temporary 3s dedupe guard', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('winFanfare'));
      act(() => vi.advanceTimersByTime(1000));
      act(() => result.current.playSound('winFanfare'));
      expect(mockPlay).toHaveBeenCalledTimes(1);
      act(() => vi.advanceTimersByTime(2100));
      act(() => result.current.playSound('winFanfare'));
      expect(mockPlay).toHaveBeenCalledTimes(2);
    });

    it('does not cross-throttle different sounds', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('cardPlace'));
      act(() => result.current.playSound('cardFlip'));
      expect(mockPlay).toHaveBeenCalledTimes(2);
    });
  });

  describe('policy layer: exec sound claim', () => {
    beforeEach(() => vi.useFakeTimers());
    afterEach(() => vi.useRealTimers());

    it('consumeExecClaim returns true once after claimExecSound, then false', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.claimExecSound());
      expect(result.current.consumeExecClaim()).toBe(true);
      expect(result.current.consumeExecClaim()).toBe(false);
    });

    it('returns false when nothing was claimed', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      expect(result.current.consumeExecClaim()).toBe(false);
    });

    it('expires an unconsumed claim after ~3s', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.claimExecSound());
      act(() => vi.advanceTimersByTime(3100));
      expect(result.current.consumeExecClaim()).toBe(false);
    });
  });

  describe('policy layer: cardPlace arpeggio', () => {
    beforeEach(() => vi.useFakeTimers());
    afterEach(() => vi.useRealTimers());

    it('ramps rate upward for consecutive cardPlace plays within 1.5s', () => {
      mockPlay.mockReturnValue(42);
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('cardPlace'));
      act(() => vi.advanceTimersByTime(200));
      act(() => result.current.playSound('cardPlace'));
      act(() => vi.advanceTimersByTime(200));
      act(() => result.current.playSound('cardPlace'));
      const rates = mockRate.mock.calls.map((c) => c[0] as number);
      expect(rates[0]).toBeCloseTo(1.0, 5);
      expect(rates[1]).toBeCloseTo(1.035, 5);
      expect(rates[2]).toBeCloseTo(1.07, 5);
    });

    it('caps the ramp at 1.25', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      for (let i = 0; i < 12; i++) {
        act(() => result.current.playSound('cardPlace'));
        act(() => vi.advanceTimersByTime(200));
      }
      const rates = mockRate.mock.calls.map((c) => c[0] as number);
      expect(Math.max(...rates)).toBeLessThanOrEqual(1.25);
    });

    it('resets to 1.0 after 1.5s of idle', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('cardPlace'));
      act(() => vi.advanceTimersByTime(200));
      act(() => result.current.playSound('cardPlace'));
      act(() => vi.advanceTimersByTime(1600));
      act(() => result.current.playSound('cardPlace'));
      const rates = mockRate.mock.calls.map((c) => c[0] as number);
      expect(rates[2]).toBeCloseTo(1.0, 5);
    });

    it('explicit pitchVariation overrides the arpeggio rate', () => {
      const { result } = renderHook(() => useSound(), { wrapper });
      act(() => result.current.playSound('cardPlace'));
      act(() => vi.advanceTimersByTime(200));
      act(() => result.current.playSound('cardPlace', { pitchVariation: 0.05 }));
      const secondRate = mockRate.mock.calls[1][0] as number;
      expect(secondRate).toBeGreaterThanOrEqual(0.95);
      expect(secondRate).toBeLessThanOrEqual(1.05);
    });
  });

  describe('policy layer: per-sound enable map', () => {
    it('skips a sound whose SOUND_ENABLED entry is false', () => {
      const original = SOUND_ENABLED.turnTick;
      SOUND_ENABLED.turnTick = false;
      try {
        const { result } = renderHook(() => useSound(), { wrapper });
        act(() => result.current.playSound('turnTick'));
        expect(mockPlay).not.toHaveBeenCalled();
        act(() => result.current.playSound('cardFlip'));
        expect(mockPlay).toHaveBeenCalledTimes(1);
      } finally {
        SOUND_ENABLED.turnTick = original;
      }
    });
  });

  describe('policy layer: suspended AudioContext gate', () => {
    it('skips plays while the context is suspended (autoUnlock would replay them stale)', () => {
      mockHowlerCtx.state = 'suspended';
      try {
        const { result } = renderHook(() => useSound(), { wrapper });
        act(() => result.current.playSound('shuffle'));
        expect(mockPlay).not.toHaveBeenCalled();
      } finally {
        mockHowlerCtx.state = 'running';
      }
    });
  });

  describe('children rendering', () => {
    it('renders children', () => {
      render(
        <SoundProvider>
          <div data-testid="child">Hello</div>
        </SoundProvider>,
      );
      expect(screen.getByTestId('child')).toBeInTheDocument();
    });
  });
});
