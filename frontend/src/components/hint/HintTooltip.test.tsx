import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { HintTooltip } from './HintTooltip';

describe('HintTooltip', () => {
  it('displays reason text', () => {
    render(<HintTooltip reason="Stand is recommended" confidence="strong" />);
    expect(screen.getByText('Stand is recommended')).toBeInTheDocument();
  });

  it('has role status for accessibility', () => {
    render(<HintTooltip reason="test" confidence="strong" />);
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('uses solid border for strong confidence', () => {
    render(<HintTooltip reason="test" confidence="strong" />);
    const el = screen.getByTestId('hint-tooltip');
    expect(el).toHaveClass('border-yellow-400');
    expect(el).not.toHaveClass('border-dashed');
  });

  it('uses dashed border for moderate confidence', () => {
    render(<HintTooltip reason="test" confidence="moderate" />);
    const el = screen.getByTestId('hint-tooltip');
    expect(el).toHaveClass('border-dashed');
  });

  it('has aria-live polite for screen readers', () => {
    render(<HintTooltip reason="test" confidence="moderate" />);
    expect(screen.getByTestId('hint-tooltip')).toHaveAttribute('aria-live', 'polite');
  });
});
