import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { DaifugoPlayerData } from '../../types/card';
import { DaifugoHumanArea } from './DaifugoHumanArea';

function makePlayer(overrides: Partial<DaifugoPlayerData> = {}): DaifugoPlayerData {
  return {
    id: 0,
    isHuman: true,
    isFinished: false,
    rank: 0,
    cardCount: 2,
    cards: [
      { design: 0, value: 1 },
      { design: 1, value: 13 },
    ],
    ...overrides,
  };
}

describe('DaifugoHumanArea', () => {
  it('renders player name "あなた" for human player', () => {
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[]} onToggle={vi.fn()} isCurrentTurn={false} onDragCard={vi.fn()} />,
    );
    expect(screen.getByText('あなた')).toBeInTheDocument();
  });

  it('shows card count and select prompt when current turn and not finished', () => {
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[]} onToggle={vi.fn()} isCurrentTurn={true} onDragCard={vi.fn()} />,
    );
    expect(screen.getByText('2枚')).toBeInTheDocument();
    expect(screen.getByText('カードをクリックして選択')).toBeInTheDocument();
  });

  it('shows card count without select prompt when not current turn and not finished', () => {
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[]} onToggle={vi.fn()} isCurrentTurn={false} onDragCard={vi.fn()} />,
    );
    expect(screen.getByText('2枚')).toBeInTheDocument();
    expect(screen.queryByText('カードをクリックして選択')).not.toBeInTheDocument();
  });

  it('does not show card count when finished (opacity 0.5 style applied)', () => {
    render(
      <DaifugoHumanArea
        player={makePlayer({ isFinished: true, rank: 1 })}
        selectedIndices={[]}
        onToggle={vi.fn()}
        isCurrentTurn={false}
        onDragCard={vi.fn()}
      />,
    );
    expect(screen.queryByText(/枚/)).not.toBeInTheDocument();
  });

  it('shows finished badge with rank when finished', () => {
    render(
      <DaifugoHumanArea
        player={makePlayer({ isFinished: true, rank: 1 })}
        selectedIndices={[]}
        onToggle={vi.fn()}
        isCurrentTurn={false}
        onDragCard={vi.fn()}
      />,
    );
    expect(screen.getByText('上がり (大富豪)')).toBeInTheDocument();
  });

  it('does not show finished badge when not finished', () => {
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[]} onToggle={vi.fn()} isCurrentTurn={false} onDragCard={vi.fn()} />,
    );
    expect(screen.queryByText(/上がり/)).not.toBeInTheDocument();
  });

  it('shows illegalFinishPenalty badge when penalty true', () => {
    render(
      <DaifugoHumanArea
        player={makePlayer({ illegalFinishPenalty: true })}
        selectedIndices={[]}
        onToggle={vi.fn()}
        isCurrentTurn={false}
        onDragCard={vi.fn()}
      />,
    );
    expect(screen.getByText('反則上がり')).toBeInTheDocument();
  });

  it('does not show illegalFinishPenalty badge when penalty falsy', () => {
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[]} onToggle={vi.fn()} isCurrentTurn={false} onDragCard={vi.fn()} />,
    );
    expect(screen.queryByText('反則上がり')).not.toBeInTheDocument();
  });

  it('cards are disabled when not current turn', () => {
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[]} onToggle={vi.fn()} isCurrentTurn={false} onDragCard={vi.fn()} />,
    );
    const buttons = screen.getAllByRole('button');
    for (const btn of buttons) {
      expect(btn).toBeDisabled();
    }
  });

  it('cards are enabled when current turn', () => {
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[]} onToggle={vi.fn()} isCurrentTurn={true} onDragCard={vi.fn()} />,
    );
    const buttons = screen.getAllByRole('button');
    for (const btn of buttons) {
      expect(btn).toBeEnabled();
    }
  });

  it('selected card has highlighted border (3px solid #f0ad4e)', () => {
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[0]} onToggle={vi.fn()} isCurrentTurn={true} onDragCard={vi.fn()} />,
    );
    const buttons = screen.getAllByRole('button');
    expect(buttons[0].style.border).toBe('3px solid rgb(240, 173, 78)');
  });

  it('unselected card has transparent border', () => {
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[0]} onToggle={vi.fn()} isCurrentTurn={true} onDragCard={vi.fn()} />,
    );
    const buttons = screen.getAllByRole('button');
    expect(buttons[1].style.border).toBe('3px solid transparent');
  });

  it('calls onToggle when card clicked', () => {
    const onToggle = vi.fn();
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[]} onToggle={onToggle} isCurrentTurn={true} onDragCard={vi.fn()} />,
    );
    fireEvent.click(screen.getAllByRole('button')[0]);
    expect(onToggle).toHaveBeenCalledWith(0);
  });

  it('calls onDragCard on dragStart', () => {
    const onDragCard = vi.fn();
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[]} onToggle={vi.fn()} isCurrentTurn={true} onDragCard={onDragCard} />,
    );
    fireEvent.dragStart(screen.getAllByRole('button')[1], {
      dataTransfer: { setData: vi.fn() },
    });
    expect(onDragCard).toHaveBeenCalledWith(1);
  });

  it('applies active turn style (border + boxShadow) when current turn', () => {
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[]} onToggle={vi.fn()} isCurrentTurn={true} onDragCard={vi.fn()} />,
    );
    const area = document.getElementById('player-area-0');
    expect(area?.style.border).toBe('2px solid rgb(92, 184, 92)');
    expect(area?.style.boxShadow).toBe('0 0 12px #5cb85c');
  });

  it('applies opacity 0.5 when finished', () => {
    render(
      <DaifugoHumanArea
        player={makePlayer({ isFinished: true, rank: 1 })}
        selectedIndices={[]}
        onToggle={vi.fn()}
        isCurrentTurn={false}
        onDragCard={vi.fn()}
      />,
    );
    const area = document.getElementById('player-area-0');
    expect(area).toHaveStyle({ opacity: 0.5 });
  });

  it('applies no conditional style when not current turn and not finished', () => {
    render(
      <DaifugoHumanArea player={makePlayer()} selectedIndices={[]} onToggle={vi.fn()} isCurrentTurn={false} onDragCard={vi.fn()} />,
    );
    const area = document.getElementById('player-area-0');
    expect(area).not.toHaveStyle({ opacity: 0.5 });
    expect(area).not.toHaveStyle({ border: '2px solid #5cb85c' });
  });

  it('renders no card buttons when cards is undefined/empty', () => {
    render(
      <DaifugoHumanArea
        player={makePlayer({ cards: undefined, cardCount: 0 })}
        selectedIndices={[]}
        onToggle={vi.fn()}
        isCurrentTurn={false}
        onDragCard={vi.fn()}
      />,
    );
    expect(screen.queryAllByRole('button')).toHaveLength(0);
  });
});
