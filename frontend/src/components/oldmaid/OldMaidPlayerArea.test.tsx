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

describe('OldMaidPlayerArea suspect/target a11y', () => {
  it('exposes the suspect toggle state via aria-pressed and a descriptive aria-label', () => {
    const { rerender } = render(
      <OldMaidPlayerArea {...defaultProps} player={makeCpuPlayer(5)} onToggleSuspect={vi.fn()} isSuspect={false} />,
    );
    const btn = screen.getByRole('button', { name: /容疑者/ });
    expect(btn).toHaveAttribute('aria-pressed', 'false');
    expect(btn.getAttribute('aria-label')).toContain('CPU');

    rerender(<OldMaidPlayerArea {...defaultProps} player={makeCpuPlayer(5)} onToggleSuspect={vi.fn()} isSuspect />);
    expect(screen.getByRole('button', { name: /容疑者/ })).toHaveAttribute('aria-pressed', 'true');
  });

  it('marks the current draw target with aria-current', () => {
    const { container } = render(
      <OldMaidPlayerArea {...defaultProps} player={makeCpuPlayer(5)} isTarget isHumanTurn />,
    );
    expect(container.querySelector('#player-area-1')).toHaveAttribute('aria-current', 'true');
  });

  it('does not set aria-current on a non-target player', () => {
    const { container } = render(<OldMaidPlayerArea {...defaultProps} player={makeCpuPlayer(5)} />);
    expect(container.querySelector('#player-area-1')).not.toHaveAttribute('aria-current');
  });
});

describe('OldMaidPlayerArea does not reveal the CPU placement trap', () => {
  const selectableProps = { ...defaultProps, isTarget: true, isHumanTurn: true };

  // cpuPlacementStrategy は「人間が引きたくなる位置に奇数札を置く」誘い込み。
  // その位置を光らせると、引く前に避けるべき札を教える案内に反転する (#5476)。
  // 以前のテストは金枠とグローが**付くこと**を要求しており、バグを固定していた。
  it('renders every CPU card with the same neutral style', () => {
    render(<OldMaidPlayerArea {...selectableProps} player={makeCpuPlayer(3)} />);
    const imgs = screen.getAllByRole('button').map((btn) => btn.querySelector('img'));
    expect(imgs.length).toBeGreaterThan(1);
    for (const img of imgs) {
      expect(img).not.toBeNull();
      expect(img).toHaveStyle({ border: '2px solid transparent' });
    }
  });

  it('leaves no card visually distinguishable from the others', () => {
    render(<OldMaidPlayerArea {...selectableProps} player={makeCpuPlayer(4)} />);
    const styles = screen.getAllByRole('button').map((btn) => btn.querySelector('img')?.getAttribute('style') ?? '');
    expect(styles.length).toBeGreaterThan(1);
    // 一枚でも違えば、それが罠の位置を指す手がかりになる。
    expect(new Set(styles).size).toBe(1);
  });

  it('applies no accent colour, lift or glow to any card', () => {
    render(<OldMaidPlayerArea {...selectableProps} player={makeCpuPlayer(4)} />);
    for (const btn of screen.getAllByRole('button')) {
      const style = btn.querySelector('img')?.getAttribute('style') ?? '';
      expect(style).not.toContain('var(--color-ds-accent)');
      expect(style).not.toContain('var(--shadow-ds-accent-glow)');
      expect(style).not.toContain('translateY');
    }
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

  function cardButtons() {
    // Select only tap-to-move buttons (wrap CardImage), ignoring suspect/control buttons.
    return screen.getAllByRole('button').filter((btn) => btn.querySelector('img[alt]:not([alt=""])') !== null);
  }

  it('ArrowRight moves focus indicator to the right', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'ArrowRight' });
    // First ArrowRight from null -> index 0
    const btns = cardButtons();
    expect(btns[0]).toHaveClass('ring-2', 'ring-ds-info');
    expect(btns[1]).not.toHaveClass('ring-2');
  });

  it('ArrowLeft moves focus indicator to the left', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    // Move right twice to index 1, then left to index 0
    fireEvent.keyDown(container, { key: 'ArrowRight' });
    fireEvent.keyDown(container, { key: 'ArrowRight' });
    fireEvent.keyDown(container, { key: 'ArrowLeft' });
    const btns = cardButtons();
    expect(btns[0]).toHaveClass('ring-2', 'ring-ds-info');
    expect(btns[1]).not.toHaveClass('ring-2');
  });

  it('ArrowLeft from null sets focus to index 0', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'ArrowLeft' });
    expect(cardButtons()[0]).toHaveClass('ring-2', 'ring-ds-info');
  });

  it('focus does not go below 0', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'ArrowRight' }); // -> 0
    fireEvent.keyDown(container, { key: 'ArrowLeft' }); // clamp at 0
    fireEvent.keyDown(container, { key: 'ArrowLeft' }); // still 0
    expect(cardButtons()[0]).toHaveClass('ring-2', 'ring-ds-info');
  });

  it('focus does not go past the last card', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    // Move right 5 times (max index is 2)
    for (let i = 0; i < 5; i++) {
      fireEvent.keyDown(container, { key: 'ArrowRight' });
    }
    const btns = cardButtons();
    expect(btns[2]).toHaveClass('ring-2', 'ring-ds-info');
    expect(btns[0]).not.toHaveClass('ring-2');
    expect(btns[1]).not.toHaveClass('ring-2');
  });

  it('Escape clears focus', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const container = screen.getByTestId('human-card-container');
    fireEvent.keyDown(container, { key: 'ArrowRight' }); // -> 0
    fireEvent.keyDown(container, { key: 'Escape' });
    for (const btn of cardButtons()) {
      expect(btn).not.toHaveClass('ring-2');
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
    for (const btn of cardButtons()) {
      expect(btn).not.toHaveClass('ring-2');
    }
  });
});

describe('OldMaidPlayerArea tap-to-move reorder', () => {
  const threeCards = [makeCard('SPADE', 1), makeCard('HEART', 5), makeCard('DIAMOND', 10)];

  function cardButtons() {
    return screen.getAllByRole('button').filter((btn) => btn.querySelector('img[alt]:not([alt=""])') !== null);
  }

  it('wraps human cards in tap-target buttons when onReorder is provided', () => {
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={vi.fn()} />);
    expect(cardButtons()).toHaveLength(3);
  });

  it('renders plain images (no buttons) when onReorder is not provided', () => {
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} />);
    // No card-wrapping buttons when reorder is disabled
    const imgs = screen.getAllByRole('img');
    expect(imgs).toHaveLength(3);
    const btns = screen.queryAllByRole('button').filter((btn) => btn.querySelector('img[alt]:not([alt=""])'));
    expect(btns).toHaveLength(0);
  });

  it('first tap selects a card and sets aria-pressed', () => {
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={vi.fn()} />);
    const btns = cardButtons();
    fireEvent.click(btns[0]);
    // After click, card 0 is selected; re-query because aria-pressed may have been added
    const updated = cardButtons();
    expect(updated[0]).toHaveAttribute('aria-pressed', 'true');
    expect(updated[1]).toHaveAttribute('aria-pressed', 'false');
  });

  it('tapping the selected card again deselects it without calling onReorder', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const btns = cardButtons();
    fireEvent.click(btns[0]);
    fireEvent.click(cardButtons()[0]);
    expect(onReorder).not.toHaveBeenCalled();
    expect(cardButtons()[0]).toHaveAttribute('aria-pressed', 'false');
  });

  it('tapping a second card moves the selected card to that index', () => {
    const onReorder = vi.fn();
    render(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />);
    const btns = cardButtons();
    fireEvent.click(btns[0]); // select index 0
    fireEvent.click(cardButtons()[2]); // move to index 2
    expect(onReorder).toHaveBeenCalledWith([1, 2, 0]);
  });

  it('does not trigger reorder when gameEndFlag is true', () => {
    const onReorder = vi.fn();
    render(
      <OldMaidPlayerArea
        {...defaultProps}
        player={makeHumanPlayer(threeCards)}
        onReorder={onReorder}
        gameEndFlag={true}
      />,
    );
    const btns = cardButtons();
    fireEvent.click(btns[0]);
    fireEvent.click(cardButtons()[2]);
    expect(onReorder).not.toHaveBeenCalled();
  });

  it('does not crash when cards shrink while a selection is active', () => {
    const { rerender } = render(
      <OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={vi.fn()} />,
    );
    // Select last card (index 2)
    fireEvent.click(cardButtons()[2]);
    expect(cardButtons()[2]).toHaveAttribute('aria-pressed', 'true');

    // Cards shrink (pair discarded) — selectedForMove is now stale (2 >= new length 1).
    // Render must not crash on the aria-label IIFE during the window before
    // the cleanup effect resets selectedForMove.
    const oneCard = [makeCard('SPADE', 1)];
    rerender(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(oneCard)} onReorder={vi.fn()} />);
    expect(cardButtons()).toHaveLength(1);
  });

  it('resets stale selection instead of reordering with an invalid index', () => {
    const onReorder = vi.fn();
    const { rerender } = render(
      <OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(threeCards)} onReorder={onReorder} />,
    );
    // Select the last index, then shrink cards so the selection is out of bounds.
    fireEvent.click(cardButtons()[2]);
    const oneCard = [makeCard('SPADE', 1)];
    rerender(<OldMaidPlayerArea {...defaultProps} player={makeHumanPlayer(oneCard)} onReorder={onReorder} />);

    // Tapping the only remaining card with a stale selectedForMove must reset,
    // not fire an onReorder call with a bogus indices array.
    fireEvent.click(cardButtons()[0]);
    expect(onReorder).not.toHaveBeenCalled();
  });
});
