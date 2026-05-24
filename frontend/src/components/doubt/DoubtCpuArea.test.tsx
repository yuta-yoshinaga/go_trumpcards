import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { DoubtPlayerData } from '../../types/card';
import { DoubtCpuArea } from './DoubtCpuArea';

const basePlayer: DoubtPlayerData = {
  id: 1,
  isHuman: false,
  cardCount: 5,
  cards: [],
  isFinished: false,
};

describe('DoubtCpuArea', () => {
  it('does not render the tell indicator when hasTell is false', () => {
    render(<DoubtCpuArea player={basePlayer} isCurrentTurn={false} hasTell={false} />);
    expect(screen.queryByTestId('doubt-tell-indicator')).not.toBeInTheDocument();
  });

  it('renders sweat, eye-dart, and label badge when hasTell is true', () => {
    render(<DoubtCpuArea player={basePlayer} isCurrentTurn={true} hasTell={true} />);
    const indicator = screen.getByTestId('doubt-tell-indicator');
    expect(indicator).toBeInTheDocument();
    expect(indicator.querySelector('.animate-sweat-drop')).not.toBeNull();
    expect(indicator.querySelector('.animate-eye-dart')).not.toBeNull();
    expect(screen.getByText('怪しい')).toBeInTheDocument();
  });
});
