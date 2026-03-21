import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { PhaseTransition } from './PhaseTransition';

vi.mock('../../hooks/useReducedMotion', () => ({
  useReducedMotion: vi.fn(() => false),
}));

import { useReducedMotion } from '../../hooks/useReducedMotion';

describe('PhaseTransition', () => {
  it('renders children with animation wrapper when motion is enabled', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(
      <PhaseTransition phaseKey="action">
        <div>Action Phase</div>
      </PhaseTransition>,
    );
    expect(screen.getByTestId('phase-transition')).toBeInTheDocument();
    expect(screen.getByText('Action Phase')).toBeInTheDocument();
  });

  it('renders children without animation wrapper when reduced motion is preferred', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(
      <PhaseTransition phaseKey="action">
        <div>Action Phase</div>
      </PhaseTransition>,
    );
    expect(screen.queryByTestId('phase-transition')).not.toBeInTheDocument();
    expect(screen.getByText('Action Phase')).toBeInTheDocument();
  });

  it('accepts numeric phase key', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(
      <PhaseTransition phaseKey={2}>
        <div>Phase 2</div>
      </PhaseTransition>,
    );
    expect(screen.getByText('Phase 2')).toBeInTheDocument();
  });
});
