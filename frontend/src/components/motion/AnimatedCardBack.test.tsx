import { render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useOptionalSound } from '../../providers/SoundProvider';
import { AnimatedCardBack } from './AnimatedCardBack';

vi.mock('../../hooks/useReducedMotion', () => ({
  useReducedMotion: vi.fn(() => false),
}));

// Override the global framer-motion mock so onAnimationComplete fires
// on mount via useEffect. See AnimatedCard.test.tsx for the rationale.
vi.mock('framer-motion', async () => {
  const React = await import('react');
  function createMotionProxy() {
    return new Proxy(
      {},
      {
        get: (_target: Record<string, unknown>, prop: string) =>
          React.forwardRef((props: Record<string, unknown>, ref: React.Ref<HTMLElement>) => {
            const {
              initial: _i,
              animate: _a,
              exit: _e,
              transition: _t,
              whileHover: _wh,
              whileTap: _wt,
              layout: _l,
              layoutId: _li,
              onAnimationComplete,
              ...rest
            } = props;
            const cb = onAnimationComplete as (() => void) | undefined;
            React.useEffect(() => {
              cb?.();
            }, [cb]);
            return React.createElement(prop, { ...rest, ref });
          }),
      },
    );
  }
  const AnimatePresence = ({ children }: { children: React.ReactNode }) =>
    React.createElement(React.Fragment, null, children);
  return { motion: createMotionProxy(), AnimatePresence };
});

const mockPlaySound = vi.fn();

// Mock SoundProvider so tests can render the component without a
// provider wrap. Mirrors AnimatedCard.test.tsx (issue #1845).
vi.mock('../../providers/SoundProvider', () => ({
  useOptionalSound: vi.fn(() => ({ playSound: mockPlaySound, muted: false, toggleMute: vi.fn() })),
}));

import { useReducedMotion } from '../../hooks/useReducedMotion';

describe('AnimatedCardBack', () => {
  beforeEach(() => {
    mockPlaySound.mockClear();
  });

  it('renders animated wrapper when motion is enabled', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<AnimatedCardBack />);
    expect(screen.getByTestId('animated-card-back')).toBeInTheDocument();
  });

  it('renders plain CardBack when reduced motion is preferred', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(<AnimatedCardBack />);
    expect(screen.queryByTestId('animated-card-back')).not.toBeInTheDocument();
    expect(screen.getByRole('img')).toBeInTheDocument();
  });

  it('passes width and style to CardBack', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(<AnimatedCardBack width={60} style={{ opacity: 0.5 }} />);
    const img = screen.getByRole('img');
    expect(img).toHaveStyle({ width: '60px', opacity: '0.5' });
  });

  it('renders as button when onClick is provided in reduced mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const onClick = vi.fn();
    render(<AnimatedCardBack onClick={onClick} ariaLabel="flip card" />);
    expect(screen.getByRole('button', { name: 'flip card' })).toBeInTheDocument();
  });

  it('renders as button when onClick is provided in animated mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onClick = vi.fn();
    render(<AnimatedCardBack onClick={onClick} ariaLabel="flip card" />);
    expect(screen.getByRole('button', { name: 'flip card' })).toBeInTheDocument();
  });

  it('applies custom dealDelay', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<AnimatedCardBack dealDelay={0.3} />);
    expect(screen.getByTestId('animated-card-back')).toBeInTheDocument();
  });

  it('accepts onFlipComplete callback prop', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onFlipComplete = vi.fn();
    render(<AnimatedCardBack onFlipComplete={onFlipComplete} />);
    expect(screen.getByTestId('animated-card-back')).toBeInTheDocument();
  });

  it('does not pass onFlipComplete in reduced motion mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const onFlipComplete = vi.fn();
    render(<AnimatedCardBack onFlipComplete={onFlipComplete} />);
    expect(screen.queryByTestId('animated-card-back')).not.toBeInTheDocument();
  });

  describe('default SFX contract (issue #1845)', () => {
    it('plays cardFlip on flip-in completion', async () => {
      vi.mocked(useReducedMotion).mockReturnValue(false);
      render(<AnimatedCardBack />);
      await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('cardFlip'));
    });

    it('also fires onFlipComplete after the default SFX', async () => {
      vi.mocked(useReducedMotion).mockReturnValue(false);
      const onFlipComplete = vi.fn();
      render(<AnimatedCardBack onFlipComplete={onFlipComplete} />);
      await waitFor(() => expect(onFlipComplete).toHaveBeenCalledTimes(1));
      expect(mockPlaySound).toHaveBeenCalledWith('cardFlip');
    });

    it('does not play sound when silent=true (but still fires onFlipComplete)', async () => {
      vi.mocked(useReducedMotion).mockReturnValue(false);
      const onFlipComplete = vi.fn();
      render(<AnimatedCardBack silent onFlipComplete={onFlipComplete} />);
      await waitFor(() => expect(onFlipComplete).toHaveBeenCalledTimes(1));
      expect(mockPlaySound).not.toHaveBeenCalled();
    });

    it('does not throw when rendered outside a SoundProvider', () => {
      vi.mocked(useOptionalSound).mockReturnValueOnce(null);
      vi.mocked(useReducedMotion).mockReturnValue(false);
      expect(() => render(<AnimatedCardBack />)).not.toThrow();
    });
  });
});
