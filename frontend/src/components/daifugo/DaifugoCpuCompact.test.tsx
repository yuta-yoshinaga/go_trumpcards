import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { DaifugoPlayerData } from '../../types/card';
import { DaifugoCpuCompact } from './DaifugoCpuCompact';

function makePlayer(overrides: Partial<DaifugoPlayerData> = {}): DaifugoPlayerData {
  return {
    id: 1,
    isHuman: false,
    isFinished: false,
    rank: 0,
    cardCount: 13,
    cards: [],
    ...overrides,
  };
}

describe('DaifugoCpuCompact', () => {
  it('renders player name and card count on one row', () => {
    render(<DaifugoCpuCompact player={makePlayer({ id: 1, cardCount: 13 })} isCurrentTurn={false} />);
    expect(screen.getByText('CPU 1')).toBeInTheDocument();
    expect(screen.getByText('13枚')).toBeInTheDocument();
  });

  it('applies current-turn highlight when it is this CPUs turn', () => {
    const { container } = render(<DaifugoCpuCompact player={makePlayer()} isCurrentTurn />);
    const root = container.firstChild as HTMLElement;
    expect(root.className).toMatch(/border-game-status-waiting/);
  });

  it('does not apply current-turn highlight when not this CPUs turn', () => {
    const { container } = render(<DaifugoCpuCompact player={makePlayer()} isCurrentTurn={false} />);
    const root = container.firstChild as HTMLElement;
    expect(root.className).not.toMatch(/border-game-status-waiting/);
  });

  it('shows finished rank label instead of card count when finished', () => {
    render(<DaifugoCpuCompact player={makePlayer({ isFinished: true, rank: 1 })} isCurrentTurn={false} />);
    expect(screen.getByText('大富豪')).toBeInTheDocument();
    expect(screen.queryByText(/枚/)).not.toBeInTheDocument();
  });

  it('dims the row when finished', () => {
    const { container } = render(
      <DaifugoCpuCompact player={makePlayer({ isFinished: true, rank: 2 })} isCurrentTurn={false} />,
    );
    const root = container.firstChild as HTMLElement;
    expect(root.className).toMatch(/opacity-50/);
  });

  it('shows illegalFinishPenalty badge when penalty is true', () => {
    render(<DaifugoCpuCompact player={makePlayer({ illegalFinishPenalty: true })} isCurrentTurn={false} />);
    expect(screen.getByText('反則上がり')).toBeInTheDocument();
  });

  it('does not show illegalFinishPenalty badge when penalty is falsy', () => {
    render(<DaifugoCpuCompact player={makePlayer()} isCurrentTurn={false} />);
    expect(screen.queryByText('反則上がり')).not.toBeInTheDocument();
  });
});
