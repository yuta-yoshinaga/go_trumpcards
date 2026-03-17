import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { HoldemEquity } from '../types/card';
import { EquityDisplay } from './EquityDisplay';

const mockEquity: HoldemEquity = {
  winProbability: 0.75,
  handOdds: [
    { handRank: 0, handName: 'High Card', probability: 0.1 },
    { handRank: 1, handName: 'One Pair', probability: 0.5 },
    { handRank: 2, handName: 'Two Pair', probability: 0.2 },
    { handRank: 3, handName: 'Three of a Kind', probability: 0.1 },
    { handRank: 4, handName: 'Straight', probability: 0.05 },
    { handRank: 5, handName: 'Flush', probability: 0.05 },
    { handRank: 6, handName: 'Full House', probability: 0.0 },
    { handRank: 7, handName: 'Four of a Kind', probability: 0.0 },
  ],
};

describe('EquityDisplay', () => {
  it('renders equity bar with correct percentage', () => {
    render(<EquityDisplay equity={mockEquity} potOdds={33.3} />);
    expect(screen.getByText(/75%/)).toBeInTheDocument();
    const bar = screen.getByTestId('equity-bar');
    expect(bar).toHaveStyle({ width: '75%' });
  });

  it('renders pot odds', () => {
    render(<EquityDisplay equity={mockEquity} potOdds={33.3} />);
    expect(screen.getByText(/33\.3%/)).toBeInTheDocument();
  });

  it('shows +EV when equity > potOdds', () => {
    render(<EquityDisplay equity={mockEquity} potOdds={33.3} />);
    const indicator = screen.getByTestId('ev-indicator');
    expect(indicator).toHaveTextContent('+EV');
    expect(indicator).toHaveClass('text-green-400');
  });

  it('shows -EV when equity <= potOdds', () => {
    const lowEquity: HoldemEquity = { ...mockEquity, winProbability: 0.2 };
    render(<EquityDisplay equity={lowEquity} potOdds={50.0} />);
    const indicator = screen.getByTestId('ev-indicator');
    expect(indicator).toHaveTextContent('-EV');
    expect(indicator).toHaveClass('text-red-400');
  });

  it('shows hand odds breakdown after toggle', () => {
    render(<EquityDisplay equity={mockEquity} potOdds={33.3} />);
    expect(screen.queryByTestId('hand-odds-table')).not.toBeInTheDocument();

    fireEvent.click(screen.getByTestId('toggle-hand-odds'));
    expect(screen.getByTestId('hand-odds-table')).toBeInTheDocument();
    expect(screen.getByText('One Pair')).toBeInTheDocument();
    expect(screen.getByText('50.0%')).toBeInTheDocument();
  });

  it('hides hand odds with 0 probability', () => {
    render(<EquityDisplay equity={mockEquity} potOdds={33.3} />);
    fireEvent.click(screen.getByTestId('toggle-hand-odds'));
    expect(screen.queryByText('Full House')).not.toBeInTheDocument();
    expect(screen.queryByText('Four of a Kind')).not.toBeInTheDocument();
  });

  it('toggles hand odds off', () => {
    render(<EquityDisplay equity={mockEquity} potOdds={33.3} />);
    fireEvent.click(screen.getByTestId('toggle-hand-odds'));
    expect(screen.getByTestId('hand-odds-table')).toBeInTheDocument();

    fireEvent.click(screen.getByTestId('toggle-hand-odds'));
    expect(screen.queryByTestId('hand-odds-table')).not.toBeInTheDocument();
  });
});
