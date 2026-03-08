import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { DaifugoPlayerData } from '../../types/card';
import { DaifugoCpuArea } from './DaifugoCpuArea';

function makePlayer(overrides: Partial<DaifugoPlayerData> = {}): DaifugoPlayerData {
  return {
    id: 1,
    isHuman: false,
    isFinished: false,
    rank: 0,
    cardCount: 5,
    cards: [],
    ...overrides,
  };
}

describe('DaifugoCpuArea', () => {
  it('renders card count when not finished', () => {
    render(<DaifugoCpuArea player={makePlayer({ cardCount: 5 })} isCurrentTurn={false} />);
    expect(screen.getByText('5枚')).toBeInTheDocument();
  });

  it('does not render card count when finished', () => {
    render(<DaifugoCpuArea player={makePlayer({ isFinished: true, rank: 1 })} isCurrentTurn={false} />);
    expect(screen.queryByText(/枚/)).not.toBeInTheDocument();
  });

  it('shows finished label with rank when finished', () => {
    render(<DaifugoCpuArea player={makePlayer({ isFinished: true, rank: 1 })} isCurrentTurn={false} />);
    expect(screen.getByText('上がり (大富豪)')).toBeInTheDocument();
  });

  it('does not show finished label when not finished', () => {
    render(<DaifugoCpuArea player={makePlayer({ isFinished: false })} isCurrentTurn={false} />);
    expect(screen.queryByText(/上がり/)).not.toBeInTheDocument();
  });

  it('shows illegalFinishPenalty badge when penalty is true', () => {
    render(<DaifugoCpuArea player={makePlayer({ illegalFinishPenalty: true })} isCurrentTurn={false} />);
    expect(screen.getByText('反則上がり')).toBeInTheDocument();
  });

  it('does not show illegalFinishPenalty badge when penalty is falsy', () => {
    render(<DaifugoCpuArea player={makePlayer()} isCurrentTurn={false} />);
    expect(screen.queryByText('反則上がり')).not.toBeInTheDocument();
  });
});
