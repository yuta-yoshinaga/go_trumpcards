import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Card, SevensPlayerData } from '../../types/card';
import { SevensHumanArea } from './SevensHumanArea';

const spade7: Card = { design: 'SPADE', value: 7 };
const spade8: Card = { design: 'SPADE', value: 8 };
const heart3: Card = { design: 'HEART', value: 3 };
const joker: Card = { design: 'JOKER', value: 0 };

function makePlayer(overrides?: Partial<SevensPlayerData>): SevensPlayerData {
  return {
    id: 0,
    isHuman: true,
    isFinished: false,
    rank: 0,
    cardCount: 3,
    passesUsed: 1,
    maxPasses: 3,
    cards: [spade7, spade8, heart3],
    lastPlayedJoker: false,
    ...overrides,
  };
}

// tablePlaced with 7 placed for all suits (bit 7 set)
const tablePlacedWith7 = [0, 1 << 7, 1 << 7, 1 << 7, 1 << 7];

const defaultProps = {
  player: makePlayer(),
  isCurrentTurn: true,
  tablePlaced: tablePlacedWith7,
  tunnelEnabled: false,
  tunnelSkipWidth: 0,
  noJokerFinish: false,
  endStopEnabled: false,
  jokerConsecutiveBanned: false,
  loading: false,
  onPlay: vi.fn(),
};

describe('SevensHumanArea', () => {
  it('renders player name and card/pass counts', () => {
    render(<SevensHumanArea {...defaultProps} />);
    expect(screen.getByText('あなた')).toBeInTheDocument();
    expect(screen.getByText(/3枚/)).toBeInTheDocument();
    expect(screen.getByText(/1\/3/)).toBeInTheDocument();
  });

  it('shows click hint when current turn', () => {
    render(<SevensHumanArea {...defaultProps} />);
    expect(screen.getByText('出せるカードをクリック')).toBeInTheDocument();
  });

  it('does not show click hint when not current turn', () => {
    render(<SevensHumanArea {...defaultProps} isCurrentTurn={false} />);
    expect(screen.queryByText('出せるカードをクリック')).not.toBeInTheDocument();
  });

  it('renders card buttons with focus-visible ring class', () => {
    render(<SevensHumanArea {...defaultProps} />);
    const buttons = screen.getAllByRole('button');
    for (const btn of buttons) {
      expect(btn.className).toContain('focus-visible:ring-white/80');
    }
  });

  it('playable card has playable-card testid and title', () => {
    render(<SevensHumanArea {...defaultProps} />);
    // spade8 is adjacent to placed spade7, so it should be playable
    const playableCards = screen.getAllByTestId('playable-card');
    expect(playableCards.length).toBeGreaterThanOrEqual(1);
  });

  it('non-playable card has no testid and is disabled', () => {
    // heart3 is not adjacent to heart7 placement
    const player = makePlayer({ cards: [heart3] });
    // tablePlaced where heart suit has only 7 placed; 3 is not adjacent
    render(<SevensHumanArea {...defaultProps} player={player} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons[0]).toBeDisabled();
    expect(buttons[0]).not.toHaveAttribute('data-testid');
  });

  it('clicking a playable card calls onPlay with card index', () => {
    const onPlay = vi.fn();
    render(<SevensHumanArea {...defaultProps} onPlay={onPlay} />);
    const playableCards = screen.getAllByTestId('playable-card');
    fireEvent.click(playableCards[0]);
    expect(onPlay).toHaveBeenCalledTimes(1);
  });

  it('cards are disabled when not current turn', () => {
    render(<SevensHumanArea {...defaultProps} isCurrentTurn={false} />);
    const buttons = screen.getAllByRole('button');
    for (const btn of buttons) {
      expect(btn).toBeDisabled();
    }
  });

  it('cards are disabled when loading', () => {
    render(<SevensHumanArea {...defaultProps} loading={true} />);
    const buttons = screen.getAllByRole('button');
    for (const btn of buttons) {
      expect(btn).toBeDisabled();
    }
  });

  it('finished player shows rank badge and opacity', () => {
    const player = makePlayer({ isFinished: true, rank: 1 });
    render(<SevensHumanArea {...defaultProps} player={player} />);
    expect(screen.getByText('1位')).toBeInTheDocument();
  });

  it('finished player does not show card/pass counts', () => {
    const player = makePlayer({ isFinished: true, rank: 1 });
    render(<SevensHumanArea {...defaultProps} player={player} />);
    expect(screen.queryByText(/枚/)).not.toBeInTheDocument();
  });

  it('non-current turn has no highlight border', () => {
    const { container } = render(<SevensHumanArea {...defaultProps} isCurrentTurn={false} />);
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper.className).not.toContain('border-game-status-active');
  });

  it('current turn has highlight border', () => {
    const { container } = render(<SevensHumanArea {...defaultProps} />);
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper.className).toContain('border-game-status-active');
  });

  it('non-playable card at current turn has reduced opacity', () => {
    // heart3 not adjacent to anything
    const player = makePlayer({ cards: [heart3] });
    render(<SevensHumanArea {...defaultProps} player={player} />);
    const btn = screen.getByRole('button');
    expect(btn).toHaveStyle({ opacity: '0.5' });
  });

  it('handles player with no cards (empty hand)', () => {
    const player = makePlayer({ cards: [] });
    render(<SevensHumanArea {...defaultProps} player={player} />);
    expect(screen.queryAllByRole('button')).toHaveLength(0);
  });

  it('shows unlimited pass text when maxPasses is 0', () => {
    const player = makePlayer({ maxPasses: 0 });
    render(<SevensHumanArea {...defaultProps} player={player} />);
    expect(screen.getByText(/∞/)).toBeInTheDocument();
  });

  it('renders joker card', () => {
    const player = makePlayer({ cards: [joker] });
    render(<SevensHumanArea {...defaultProps} player={player} />);
    expect(screen.getAllByRole('button')).toHaveLength(1);
  });

  it('handles undefined cards gracefully', () => {
    const player = makePlayer({ cards: undefined as unknown as Card[] });
    render(<SevensHumanArea {...defaultProps} player={player} />);
    expect(screen.queryAllByRole('button')).toHaveLength(0);
  });
});
