import { act, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { PartnerRevealFlash } from './PartnerRevealFlash';

describe('PartnerRevealFlash', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders nothing when revealed is false', () => {
    render(<PartnerRevealFlash revealed={false} partnerName="P2" headline="Adjutant!" />);
    expect(screen.queryByTestId('partner-reveal-flash')).not.toBeInTheDocument();
  });

  it('appears when revealed transitions from false to true', () => {
    const { rerender } = render(<PartnerRevealFlash revealed={false} partnerName="P2" headline="Adjutant!" />);
    rerender(<PartnerRevealFlash revealed={true} partnerName="P2" headline="Adjutant!" />);
    expect(screen.getByTestId('partner-reveal-flash')).toBeInTheDocument();
    expect(screen.getByText('P2')).toBeInTheDocument();
    expect(screen.getByText('Adjutant!')).toBeInTheDocument();
  });

  it('auto-dismisses after the flash timeout', () => {
    const { rerender } = render(<PartnerRevealFlash revealed={false} partnerName="P2" headline="Adjutant!" />);
    rerender(<PartnerRevealFlash revealed={true} partnerName="P2" headline="Adjutant!" />);
    expect(screen.getByTestId('partner-reveal-flash')).toBeInTheDocument();
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(screen.queryByTestId('partner-reveal-flash')).not.toBeInTheDocument();
  });

  it('re-arms when revealed flips back to false', () => {
    const { rerender } = render(<PartnerRevealFlash revealed={false} partnerName="P2" headline="Adjutant!" />);
    rerender(<PartnerRevealFlash revealed={true} partnerName="P2" headline="Adjutant!" />);
    expect(screen.getByTestId('partner-reveal-flash')).toBeInTheDocument();

    rerender(<PartnerRevealFlash revealed={false} partnerName="P2" headline="Adjutant!" />);
    expect(screen.queryByTestId('partner-reveal-flash')).not.toBeInTheDocument();

    rerender(<PartnerRevealFlash revealed={true} partnerName="P2" headline="Adjutant!" />);
    expect(screen.getByTestId('partner-reveal-flash')).toBeInTheDocument();
  });
});
