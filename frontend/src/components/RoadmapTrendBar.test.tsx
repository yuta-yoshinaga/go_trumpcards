import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { RoadmapTrendBar } from './RoadmapTrendBar';

const baseProps = {
  leftCode: 1,
  rightCode: 2,
  leftLabel: 'P',
  rightLabel: 'B',
};

describe('RoadmapTrendBar', () => {
  it('returns null for empty history', () => {
    const { container } = render(<RoadmapTrendBar history={[]} {...baseProps} />);
    expect(container.firstChild).toBeNull();
  });

  it('shows a streak badge with fire when 3+ in a row', () => {
    render(<RoadmapTrendBar history={[2, 1, 1, 1]} {...baseProps} />);
    expect(screen.getByTestId('roadmap-trend-streak')).toBeInTheDocument();
  });

  it('does not show fire badge when streak is < 3', () => {
    render(<RoadmapTrendBar history={[1, 2, 2]} {...baseProps} />);
    expect(screen.queryByTestId('roadmap-trend-streak')).not.toBeInTheDocument();
  });

  it('ignores neutral outcomes when computing streaks', () => {
    // Ties (code 0) shouldn't break the run of 1s.
    render(<RoadmapTrendBar history={[1, 0, 1, 0, 1]} {...baseProps} />);
    expect(screen.getByTestId('roadmap-trend-streak')).toBeInTheDocument();
  });

  it('renders percentage labels for both sides', () => {
    render(<RoadmapTrendBar history={[1, 1, 2, 2]} {...baseProps} />);
    expect(screen.getByText('P 50%')).toBeInTheDocument();
    expect(screen.getByText('50% B')).toBeInTheDocument();
  });
});
