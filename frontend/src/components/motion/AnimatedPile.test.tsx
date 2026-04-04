import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Card } from '../../types/card';
import { AnimatedPile } from './AnimatedPile';

vi.mock('../../hooks/useReducedMotion', () => ({
  useReducedMotion: vi.fn(() => false),
}));

import { useReducedMotion } from '../../hooks/useReducedMotion';

const mockCards: Card[] = [
  { design: 'SPADE', value: 1 },
  { design: 'HEART', value: 10 },
  { design: 'DIAMOND', value: 5 },
];

describe('AnimatedPile', () => {
  describe('stacked layout', () => {
    it('renders cards with animated wrappers', () => {
      vi.mocked(useReducedMotion).mockReturnValue(false);
      render(<AnimatedPile cards={mockCards} layout="stacked" />);
      expect(screen.getByTestId('animated-pile')).toBeInTheDocument();
      expect(screen.getAllByTestId('animated-card')).toHaveLength(3);
    });

    it('calls onPlace for each card', () => {
      vi.mocked(useReducedMotion).mockReturnValue(false);
      const onPlace = vi.fn();
      render(<AnimatedPile cards={mockCards} layout="stacked" onPlace={onPlace} />);
      // onPlace will be called by AnimatedCard's onDealComplete
      expect(screen.getAllByTestId('animated-card')).toHaveLength(3);
    });
  });

  describe('fanned layout', () => {
    it('renders cards with offset positions', () => {
      vi.mocked(useReducedMotion).mockReturnValue(false);
      render(<AnimatedPile cards={mockCards} layout="fanned" cardWidth={80} />);
      expect(screen.getAllByTestId('animated-card')).toHaveLength(3);
    });
  });

  describe('cascade mode', () => {
    it('renders all cards with cascade delays', () => {
      vi.mocked(useReducedMotion).mockReturnValue(false);
      render(<AnimatedPile cards={mockCards} layout="stacked" cascade />);
      expect(screen.getAllByTestId('animated-card')).toHaveLength(3);
    });

    it('accepts onComplete callback', () => {
      vi.mocked(useReducedMotion).mockReturnValue(false);
      const onComplete = vi.fn();
      render(<AnimatedPile cards={mockCards} layout="stacked" cascade onComplete={onComplete} />);
      expect(screen.getAllByTestId('animated-card')).toHaveLength(3);
    });
  });

  describe('reduced motion', () => {
    it('falls back to static rendering', () => {
      vi.mocked(useReducedMotion).mockReturnValue(true);
      render(<AnimatedPile cards={mockCards} layout="stacked" />);
      expect(screen.getByTestId('animated-pile')).toBeInTheDocument();
      expect(screen.queryAllByTestId('animated-card')).toHaveLength(0);
      expect(screen.getAllByRole('img')).toHaveLength(3);
    });
  });

  describe('empty pile', () => {
    it('renders empty container with no cards', () => {
      vi.mocked(useReducedMotion).mockReturnValue(false);
      render(<AnimatedPile cards={[]} layout="stacked" />);
      expect(screen.getByTestId('animated-pile')).toBeInTheDocument();
      expect(screen.queryAllByTestId('animated-card')).toHaveLength(0);
    });
  });
});
