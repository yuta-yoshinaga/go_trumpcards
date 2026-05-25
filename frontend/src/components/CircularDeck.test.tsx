import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { CircularDeck } from './CircularDeck';

describe('CircularDeck', () => {
  it('renders the empty placeholder when no cards remain', () => {
    render(<CircularDeck count={0} cardWidth={32} onDrawCard={vi.fn()} drawAriaLabel="draw" />);
    expect(screen.getByTestId('circular-deck-empty')).toBeInTheDocument();
  });

  it('renders one button per card up to the cap', () => {
    render(<CircularDeck count={5} cardWidth={32} onDrawCard={vi.fn()} drawAriaLabel="draw" />);
    expect(screen.getAllByRole('button')).toHaveLength(5);
  });

  it('caps the number of rendered cards regardless of count', () => {
    render(<CircularDeck count={100} cardWidth={32} onDrawCard={vi.fn()} drawAriaLabel="draw" />);
    // MAX_VISIBLE_CARDS = 26
    expect(screen.getAllByRole('button')).toHaveLength(26);
  });

  it('fires onDrawCard when any card is tapped', () => {
    const onDrawCard = vi.fn();
    render(<CircularDeck count={4} cardWidth={32} onDrawCard={onDrawCard} drawAriaLabel="draw" />);
    fireEvent.click(screen.getAllByRole('button')[2]);
    expect(onDrawCard).toHaveBeenCalledTimes(1);
  });

  it('does not fire onDrawCard while disabled', () => {
    const onDrawCard = vi.fn();
    render(<CircularDeck count={3} cardWidth={32} onDrawCard={onDrawCard} drawAriaLabel="draw" disabled />);
    fireEvent.click(screen.getAllByRole('button')[0]);
    expect(onDrawCard).not.toHaveBeenCalled();
  });
});
