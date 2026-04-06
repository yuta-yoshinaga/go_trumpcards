import { act, render, renderHook, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { SoundProvider, useSound } from './SoundProvider';

const { mockPlay, mockRate, mockVolume } = vi.hoisted(() => ({
  mockPlay: vi.fn().mockReturnValue(1),
  mockRate: vi.fn(),
  mockVolume: vi.fn(),
}));

vi.mock('howler', () => ({
  Howl: class MockHowl {
    play = mockPlay;
    rate = mockRate;
    volume = mockVolume;
  },
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
