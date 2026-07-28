import { render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useOptionalSound } from '../../providers/SoundProvider';
import type { Card } from '../../types/card';
import { AnimatedCard } from './AnimatedCard';

// Override the global framer-motion mock from `src/test/setup.ts` so
// onAnimationComplete fires on mount via useEffect. The default mock
// strips animation props entirely, which would mean the deal callback
// — and therefore the default SFX path this file is verifying — never
// runs. Mirrors `AnimatedCardBack.test.tsx`.
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

const mockCard: Card = { design: 'SPADE', value: 1 };

vi.mock('../../hooks/useReducedMotion', () => ({
  useReducedMotion: vi.fn(() => false),
}));

import { useReducedMotion } from '../../hooks/useReducedMotion';

const mockPlaySound = vi.fn();

// Mock the SoundProvider module so tests can render AnimatedCard
// without wrapping in <SoundProvider>. The mocked useOptionalSound
// returns a spyable playSound by default; individual tests override
// it via vi.mocked(useOptionalSound).mockReturnValueOnce(null) to
// exercise the provider-less code path. See issue #1845.
vi.mock('../../providers/SoundProvider', () => ({
  useOptionalSound: vi.fn(() => ({ playSound: mockPlaySound, muted: false, toggleMute: vi.fn() })),
}));

describe('AnimatedCard', () => {
  beforeEach(() => {
    mockPlaySound.mockClear();
  });

  it('renders animated wrapper when motion is enabled', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<AnimatedCard card={mockCard} />);
    expect(screen.getByTestId('animated-card')).toBeInTheDocument();
  });

  // Reduced motion means no movement, not no feedback. This branch used to render a
  // bare CardImage, which left those users with no hover affordance on cards at all
  // while everyone else got a lift (#930). It now renders the card with a static
  // brightness/shadow hover and NO transform.
  it('renders the card with a static hover affordance and no motion when reduced motion is preferred', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const { container } = render(<AnimatedCard card={mockCard} />);
    expect(screen.getByRole('img')).toBeInTheDocument();
    const hoverTarget = container.querySelector('[class*="hover:brightness"]');
    expect(hoverTarget).not.toBeNull();
    expect(hoverTarget?.className).toContain('hover:shadow-lg');
    // No lift/scale: those are motion, which is exactly what this branch avoids.
    expect(container.innerHTML).not.toContain('translate');
    expect(container.innerHTML).not.toContain('scale(');
  });

  it('passes width and style to CardImage', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(<AnimatedCard card={mockCard} width={60} style={{ opacity: 0.5 }} />);
    const img = screen.getByRole('img');
    expect(img).toHaveStyle({ width: '60px', opacity: '0.5' });
  });

  it('passes className to CardImage', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(<AnimatedCard card={mockCard} className="test-class" />);
    expect(screen.getByRole('img')).toHaveClass('test-class');
  });

  it('applies dealDelay and isSelected defaults', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<AnimatedCard card={mockCard} />);
    expect(screen.getByTestId('animated-card')).toBeInTheDocument();
  });

  it('applies custom dealDelay and isSelected', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<AnimatedCard card={mockCard} dealDelay={0.2} isSelected={true} />);
    expect(screen.getByTestId('animated-card')).toBeInTheDocument();
  });

  it('passes drag props to CardImage in reduced mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const onDragStart = vi.fn();
    render(<AnimatedCard card={mockCard} draggable onDragStart={onDragStart} />);
    expect(screen.getByRole('img')).toHaveAttribute('draggable', 'true');
  });

  it('passes drag props to CardImage in animated mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onDragStart = vi.fn();
    render(<AnimatedCard card={mockCard} draggable onDragStart={onDragStart} />);
    expect(screen.getByRole('img')).toHaveAttribute('draggable', 'true');
  });

  it('accepts onDealComplete callback prop', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onDealComplete = vi.fn();
    render(<AnimatedCard card={mockCard} onDealComplete={onDealComplete} />);
    expect(screen.getByTestId('animated-card')).toBeInTheDocument();
  });

  it('honours a caller-supplied wrapperClassName in reduced motion mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(<AnimatedCard card={mockCard} wrapperClassName="w-16 shrink-0" />);
    const wrapper = screen.getByTestId('animated-card');
    expect(wrapper.className).toContain('w-16');
    // The default inline-block is only a fallback; a caller's layout classes win.
    expect(wrapper.className).not.toContain('inline-block');
  });

  it('never fires the deal callback or sound in reduced motion mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const onDealComplete = vi.fn();
    render(<AnimatedCard card={mockCard} onDealComplete={onDealComplete} />);
    // There is no deal animation to complete, so the callback must not fire.
    expect(onDealComplete).not.toHaveBeenCalled();
  });

  describe('default SFX contract (issue #1845)', () => {
    it('plays cardDeal on deal-in completion', async () => {
      vi.mocked(useReducedMotion).mockReturnValue(false);
      render(<AnimatedCard card={mockCard} />);
      await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('cardDeal', { pitchVariation: 0.03 }));
    });

    it('also fires onDealComplete after the default SFX', async () => {
      vi.mocked(useReducedMotion).mockReturnValue(false);
      const onDealComplete = vi.fn();
      render(<AnimatedCard card={mockCard} onDealComplete={onDealComplete} />);
      await waitFor(() => expect(onDealComplete).toHaveBeenCalledTimes(1));
      expect(mockPlaySound).toHaveBeenCalledWith('cardDeal', { pitchVariation: 0.03 });
    });

    it('does not play sound when silent=true (but still fires onDealComplete)', async () => {
      vi.mocked(useReducedMotion).mockReturnValue(false);
      const onDealComplete = vi.fn();
      render(<AnimatedCard card={mockCard} silent onDealComplete={onDealComplete} />);
      // onDealComplete is the side-effect hook callers can still opt into;
      // wait for it to confirm the animation cycle ran. Then assert no SFX.
      await waitFor(() => expect(onDealComplete).toHaveBeenCalledTimes(1));
      expect(mockPlaySound).not.toHaveBeenCalled();
    });

    it('does not throw when rendered outside a SoundProvider', () => {
      // Force useOptionalSound to behave as if no provider is mounted.
      // The component must degrade silently rather than crash.
      vi.mocked(useOptionalSound).mockReturnValueOnce(null);
      vi.mocked(useReducedMotion).mockReturnValue(false);
      expect(() => render(<AnimatedCard card={mockCard} />)).not.toThrow();
    });
  });
});
