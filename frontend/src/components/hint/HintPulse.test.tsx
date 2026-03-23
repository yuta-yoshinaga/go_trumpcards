import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { HintPulse } from './HintPulse';

describe('HintPulse', () => {
  it('renders children without wrapper when inactive', () => {
    render(
      <HintPulse active={false} reducedMotion={false}>
        <button type="button">Hit</button>
      </HintPulse>,
    );
    expect(screen.getByRole('button', { name: 'Hit' })).toBeInTheDocument();
    expect(screen.queryByTestId('hint-pulse')).not.toBeInTheDocument();
  });

  it('renders pulse ring when active and motion allowed', () => {
    render(
      <HintPulse active={true} reducedMotion={false}>
        <button type="button">Hit</button>
      </HintPulse>,
    );
    expect(screen.getByTestId('hint-pulse')).toBeInTheDocument();
    expect(screen.getByTestId('hint-ring')).toBeInTheDocument();
    expect(screen.getByTestId('hint-ring')).toHaveClass('animate-pulse');
  });

  it('renders icon instead of animation when reduced motion', () => {
    render(
      <HintPulse active={true} reducedMotion={true}>
        <button type="button">Hit</button>
      </HintPulse>,
    );
    expect(screen.getByTestId('hint-pulse')).toBeInTheDocument();
    expect(screen.getByTestId('hint-icon')).toBeInTheDocument();
    expect(screen.queryByTestId('hint-ring')).not.toBeInTheDocument();
  });

  it('children are always rendered when active', () => {
    render(
      <HintPulse active={true} reducedMotion={false}>
        <button type="button">Hit</button>
      </HintPulse>,
    );
    expect(screen.getByRole('button', { name: 'Hit' })).toBeInTheDocument();
  });
});
