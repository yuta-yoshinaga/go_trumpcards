import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../test/renderWithProviders';
import type { GoFishPlayerData } from '../../types/card';
import { GoFishPlayerArea } from './GoFishPlayerArea';

const cpuPlayer: GoFishPlayerData = {
  id: 1,
  isHuman: false,
  cardCount: 5,
  cards: [],
  bookCount: 2,
  books: [],
};

describe('GoFishPlayerArea', () => {
  it('renders player name, card count, and book count', () => {
    renderWithProviders(<GoFishPlayerArea player={cpuPlayer} isSelected={false} onSelect={vi.fn()} disabled={false} />);
    expect(screen.getByRole('button')).toBeInTheDocument();
    expect(screen.getByText(/CPU/)).toBeInTheDocument();
  });

  it('calls onSelect with player id when clicked', () => {
    const onSelect = vi.fn();
    renderWithProviders(
      <GoFishPlayerArea player={cpuPlayer} isSelected={false} onSelect={onSelect} disabled={false} />,
    );
    fireEvent.click(screen.getByRole('button'));
    expect(onSelect).toHaveBeenCalledWith(1);
  });

  it('does not call onSelect when disabled', () => {
    const onSelect = vi.fn();
    renderWithProviders(<GoFishPlayerArea player={cpuPlayer} isSelected={false} onSelect={onSelect} disabled={true} />);
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('sets aria-pressed when selected', () => {
    renderWithProviders(<GoFishPlayerArea player={cpuPlayer} isSelected={true} onSelect={vi.fn()} disabled={false} />);
    expect(screen.getByRole('button')).toHaveAttribute('aria-pressed', 'true');
  });

  it('sets aria-pressed false when not selected', () => {
    renderWithProviders(<GoFishPlayerArea player={cpuPlayer} isSelected={false} onSelect={vi.fn()} disabled={false} />);
    expect(screen.getByRole('button')).toHaveAttribute('aria-pressed', 'false');
  });

  it('does not render an ask bubble when askAnnotation is absent', () => {
    renderWithProviders(<GoFishPlayerArea player={cpuPlayer} isSelected={false} onSelect={vi.fn()} disabled={false} />);
    expect(screen.queryByTestId('cpu-action-bubble')).not.toBeInTheDocument();
  });

  it('renders hit bubble with count when askAnnotation.receivedCount > 0', () => {
    renderWithProviders(
      <GoFishPlayerArea
        player={cpuPlayer}
        isSelected={false}
        onSelect={vi.fn()}
        disabled={false}
        askAnnotation={{ rank: 5, receivedCount: 2, triggerKey: 't-1' }}
      />,
    );
    const bubble = screen.getByTestId('cpu-action-bubble');
    // "5 → ..." or "5 → 2枚！" — must contain both rank and count tokens.
    expect(bubble.textContent).toMatch(/5/);
    expect(bubble.textContent).toMatch(/2/);
  });

  it('renders miss bubble with Go Fish when receivedCount is 0', () => {
    renderWithProviders(
      <GoFishPlayerArea
        player={cpuPlayer}
        isSelected={false}
        onSelect={vi.fn()}
        disabled={false}
        askAnnotation={{ rank: 11, receivedCount: 0, triggerKey: 't-2' }}
      />,
    );
    const bubble = screen.getByTestId('cpu-action-bubble');
    // 11 → 'J' via valueName
    expect(bubble.textContent).toMatch(/J/);
    expect(bubble.textContent).toMatch(/Go Fish/);
  });
});
