import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Card, OldMaidPlayerData } from '../../types/card';
import { OldMaidPlayerArea } from './OldMaidPlayerArea';

function makeCard(design: Card['design'], value: number): Card {
  return { design, value };
}

function makeHumanPlayer(cards: Card[]): OldMaidPlayerData {
  return {
    id: 0,
    isHuman: true,
    isFinished: false,
    cardCount: cards.length,
    cards,
  };
}

function makeCpuPlayer(cardCount: number): OldMaidPlayerData {
  return {
    id: 1,
    isHuman: false,
    isFinished: false,
    cardCount,
    cards: [],
  };
}

const defaultProps = {
  isTarget: false,
  isHumanTurn: false,
  gameEndFlag: false,
  loading: false,
  highlightedCardIdx: -1,
  onDraw: vi.fn(),
};

describe('OldMaidPlayerArea compactNonTarget', () => {
  it('hides card backs for non-target CPU when compactNonTarget is true', () => {
    render(<OldMaidPlayerArea {...defaultProps} player={makeCpuPlayer(5)} compactNonTarget />);
    expect(screen.queryByAltText('カード裏面')).not.toBeInTheDocument();
    expect(screen.getByText('5枚')).toBeInTheDocument();
  });

  it('shows card backs for target CPU even when compactNonTarget is true', () => {
    render(<OldMaidPlayerArea {...defaultProps} player={makeCpuPlayer(5)} isTarget compactNonTarget />);
    const backs = screen.getAllByAltText('カード裏面');
    expect(backs.length).toBeGreaterThanOrEqual(1);
  });

  it('shows card backs for non-target CPU when compactNonTarget is false', () => {
    render(<OldMaidPlayerArea {...defaultProps} player={makeCpuPlayer(5)} />);
    const backs = screen.getAllByAltText('カード裏面');
    expect(backs).toHaveLength(5);
  });
});

describe('OldMaidPlayerArea keyboard reordering', () => {
  const threeCards = [makeCard('SPADE', 1), makeCard('HEART', 5), makeCard('DIAMOND', 10)];

  it('adds tabIndex and data-testid to human card container when onReorder is provided', () => {
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={vi.fn()} />);
    const container = screen.getByTestId('human-card-container');
    expect(container).toHaveAttribute('tabindex', '0');
  });

  it('does not add keyboard handler when onReorder is not provided', () => {
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} />);
    expect(screen.queryByTestId('human-card-container')).toBeNull();
  });

  it('ArrowRight moves focus indicator to the right', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'ArrowRight' });
    // First ArrowRight from null -> index 0
    const images = screen.getAllByRole('img');
    expect(images[0]).toHaveClass('ring-2', 'ring-blue-500');
    expect(images[1]).not.toHaveClass('ring-2');
  });

  it('ArrowLeft moves focus indicator to the left', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    // Move right twice to index 1, then left to index 0
    fireEvent.keyDown(container, { key: 'ArrowRight' });
    fireEvent.keyDown(container, { key: 'ArrowRight' });
    fireEvent.keyDown(container, { key: 'ArrowLeft' });
    const images = screen.getAllByRole('img');
    expect(images[0]).toHaveClass('ring-2', 'ring-blue-500');
    expect(images[1]).not.toHaveClass('ring-2');
  });

  it('ArrowLeft from null sets focus to index 0', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'ArrowLeft' });
    const images = screen.getAllByRole('img');
    expect(images[0]).toHaveClass('ring-2', 'ring-blue-500');
  });

  it('focus does not go below 0', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'ArrowRight' }); // -> 0
    fireEvent.keyDown(container, { key: 'ArrowLeft' }); // clamp at 0
    fireEvent.keyDown(container, { key: 'ArrowLeft' }); // still 0
    const images = screen.getAllByRole('img');
    expect(images[0]).toHaveClass('ring-2', 'ring-blue-500');
  });

  it('focus does not go past the last card', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    // Move right 5 times (max index is 2)
    for (let i = 0; i < 5; i++) {
      fireEvent.keyDown(container, { key: 'ArrowRight' });
    }
    const images = screen.getAllByRole('img');
    expect(images[2]).toHaveClass('ring-2', 'ring-blue-500');
    expect(images[0]).not.toHaveClass('ring-2');
    expect(images[1]).not.toHaveClass('ring-2');
  });

  it('Escape clears focus', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'ArrowRight' }); // -> 0
    fireEvent.keyDown(container, { key: 'Escape' });
    const images = screen.getAllByRole('img');
    for (const img of images) {
      expect(img).not.toHaveClass('ring-2');
    }
  });

  it('Shift+ArrowRight swaps focused card with right neighbor', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'ArrowRight' }); // focus -> 0
    fireEvent.keyDown(container, { key: 'ArrowRight', shiftKey: true }); // swap 0<->1, focus -> 1
    expect(onReorder).toHaveBeenCalledWith([1, 0, 2]);
  });

  it('Shift+ArrowLeft swaps focused card with left neighbor', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    // Move to index 1
    fireEvent.keyDown(container, { key: 'ArrowRight' }); // -> 0
    fireEvent.keyDown(container, { key: 'ArrowRight' }); // -> 1
    fireEvent.keyDown(container, { key: 'ArrowLeft', shiftKey: true }); // swap 1<->0, focus -> 0
    expect(onReorder).toHaveBeenCalledWith([1, 0, 2]);
  });

  it('Shift+ArrowRight does nothing when focus is on last card', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    // Move to last card
    for (let i = 0; i < 3; i++) {
      fireEvent.keyDown(container, { key: 'ArrowRight' });
    }
    fireEvent.keyDown(container, { key: 'ArrowRight', shiftKey: true });
    expect(onReorder).not.toHaveBeenCalled();
  });

  it('Shift+ArrowLeft does nothing when focus is on first card', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'ArrowRight' }); // -> 0
    fireEvent.keyDown(container, { key: 'ArrowLeft', shiftKey: true });
    expect(onReorder).not.toHaveBeenCalled();
  });

  it('Shift+ArrowLeft does nothing when focus is null', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'ArrowLeft', shiftKey: true });
    expect(onReorder).not.toHaveBeenCalled();
  });

  it('Shift+ArrowRight does nothing when focus is null', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'ArrowRight', shiftKey: true });
    expect(onReorder).not.toHaveBeenCalled();
  });

  it('does not add keyboard handler when player is finished', () => {
    const finishedPlayer: OldMaidPlayerData = {
      ...makeHumanPlayer(threeCards),
      isFinished: true,
    };
    render(<OldMaidPlayerArea {...defaultProps} player={finishedPlayer} onReorder={vi.fn()} />);
    expect(screen.queryByTestId('human-card-container')).toBeNull();
  });

  it('ignores unrelated keys', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'Enter' });
    fireEvent.keyDown(container, { key: 'ArrowUp' });
    expect(onReorder).not.toHaveBeenCalled();
    // No focus should be set
    const images = screen.getAllByRole('img');
    for (const img of images) {
      expect(img).not.toHaveClass('ring-2');
    }
  });
});
