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

  it('does not render a visible ask bubble when askAnnotation is absent', () => {
    renderWithProviders(<GoFishPlayerArea player={cpuPlayer} isSelected={false} onSelect={vi.fn()} disabled={false} />);
    // Visible bubble must be absent…
    expect(screen.queryByTestId('cpu-action-bubble')).not.toBeInTheDocument();
    // …but the sr-only live region stays mounted so announcements aren't
    // missed between mounts (see PR #1498 review).
    expect(screen.getByTestId('cpu-action-bubble-live')).toBeInTheDocument();
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

  it('omits the known-ranks chip row when knownRanks is absent or empty', () => {
    const { rerender } = renderWithProviders(
      <GoFishPlayerArea player={cpuPlayer} isSelected={false} onSelect={vi.fn()} disabled={false} />,
    );
    expect(screen.queryByTestId('known-ranks-1')).not.toBeInTheDocument();

    rerender(
      <GoFishPlayerArea player={cpuPlayer} isSelected={false} onSelect={vi.fn()} disabled={false} knownRanks={[]} />,
    );
    expect(screen.queryByTestId('known-ranks-1')).not.toBeInTheDocument();
  });

  it('renders one chip per known rank with the localized rank label', () => {
    renderWithProviders(
      <GoFishPlayerArea
        player={cpuPlayer}
        isSelected={false}
        onSelect={vi.fn()}
        disabled={false}
        knownRanks={[7, 12]}
      />,
    );
    const row = screen.getByTestId('known-ranks-1');
    const chips = row.querySelectorAll('[data-rank]');
    expect(chips).toHaveLength(2);
    // valueName(7) → '7', valueName(12) → 'Q'
    expect(row.textContent).toContain('7');
    expect(row.textContent).toContain('Q');
  });

  it('flags matched ranks with data-matched while leaving unmatched chips plain', () => {
    renderWithProviders(
      <GoFishPlayerArea
        player={cpuPlayer}
        isSelected={false}
        onSelect={vi.fn()}
        disabled={false}
        knownRanks={[7, 12]}
        matchedRanks={[7]}
      />,
    );
    const row = screen.getByTestId('known-ranks-1');
    const seven = row.querySelector('[data-rank="7"]');
    const queen = row.querySelector('[data-rank="12"]');
    expect(seven).toHaveAttribute('data-matched', 'true');
    expect(seven?.className).toContain('bg-ds-warning');
    expect(queen).not.toHaveAttribute('data-matched');
    expect(queen?.className).not.toContain('bg-ds-warning');
  });

  it('omits the book-ranks chip row when the player has no completed books', () => {
    renderWithProviders(<GoFishPlayerArea player={cpuPlayer} isSelected={false} onSelect={vi.fn()} disabled={false} />);
    expect(screen.queryByTestId('book-ranks-1')).not.toBeInTheDocument();
  });

  it("renders each of a CPU's completed book ranks with the localized rank label", () => {
    const withBooks: GoFishPlayerData = {
      ...cpuPlayer,
      books: [
        { rank: 1, cards: [] },
        { rank: 13, cards: [] },
      ],
    };
    renderWithProviders(<GoFishPlayerArea player={withBooks} isSelected={false} onSelect={vi.fn()} disabled={false} />);
    const row = screen.getByTestId('book-ranks-1');
    const chips = row.querySelectorAll('[data-rank]');
    expect(chips).toHaveLength(2);
    // valueName(1) → 'A', valueName(13) → 'K'
    expect(row.textContent).toContain('A');
    expect(row.textContent).toContain('K');
  });
});
