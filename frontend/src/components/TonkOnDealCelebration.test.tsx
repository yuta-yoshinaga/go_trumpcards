import { act, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { calcTonkHandTotal, TonkOnDealCelebration } from './TonkOnDealCelebration';

const card = (design: CardDesign, value: number): Card => ({ design, value });

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('calcTonkHandTotal', () => {
  it('uses Gin-Rummy values: A=1, 2-9 face, 10/J/Q/K=10', () => {
    const hand = [card('SPADE', 1), card('HEART', 7), card('DIAMOND', 13), card('CLOVER', 11), card('SPADE', 11)];
    // 1 + 7 + 10 + 10 + 10 = 38 (not Tonk, but the helper is unit-tested)
    expect(calcTonkHandTotal(hand)).toBe(38);
  });

  it('returns 50 for the classic Tonk-on-deal high case (5 face cards)', () => {
    const hand = [card('SPADE', 13), card('HEART', 12), card('DIAMOND', 11), card('CLOVER', 13), card('SPADE', 12)];
    expect(calcTonkHandTotal(hand)).toBe(50);
  });

  it('returns 49 for the Tonk-on-deal low case (4 face cards + 9)', () => {
    const hand = [card('SPADE', 13), card('HEART', 12), card('DIAMOND', 11), card('CLOVER', 13), card('SPADE', 9)];
    expect(calcTonkHandTotal(hand)).toBe(49);
  });

  it('treats Joker as 0 to be defensive against unexpected input', () => {
    const hand = [card('JOKER', 0)];
    expect(calcTonkHandTotal(hand)).toBe(0);
  });
});

describe('TonkOnDealCelebration', () => {
  it('keeps the aria-live container in the DOM but hidden when show=false', () => {
    render(<TonkOnDealCelebration show={false} />);
    // Container is always rendered so screen readers see a stable live region
    // (announcements inside a region added on the same tick are often dropped).
    const container = screen.getByTestId('tonk-on-deal-celebration');
    expect(container).toBeInTheDocument();
    expect(container).toHaveAttribute('data-visible', 'false');
    expect(container).toHaveClass('invisible');
    expect(screen.queryByText('TONK!')).not.toBeInTheDocument();
  });

  it('renders the TONK! banner when show=true', () => {
    render(
      <TonkOnDealCelebration
        show={true}
        winnerCards={[card('SPADE', 13), card('HEART', 12), card('DIAMOND', 11), card('CLOVER', 13), card('SPADE', 12)]}
        winnerName="あなた"
      />,
    );
    const container = screen.getByTestId('tonk-on-deal-celebration');
    expect(container).toHaveAttribute('data-visible', 'true');
    expect(container).not.toHaveClass('invisible');
    expect(screen.getByText('TONK!')).toBeInTheDocument();
    expect(screen.getByText(/50/)).toBeInTheDocument();
    expect(screen.getByText(/あなた/)).toBeInTheDocument();
  });

  it('auto-dismisses after the configured delay', () => {
    render(<TonkOnDealCelebration show={true} dismissAfterMs={1500} winnerCards={[card('SPADE', 13)]} />);
    expect(screen.getByTestId('tonk-on-deal-celebration')).toHaveAttribute('data-visible', 'true');
    act(() => {
      vi.advanceTimersByTime(1500);
    });
    // The live region remains in the DOM but is now hidden and the inner banner is gone.
    const container = screen.getByTestId('tonk-on-deal-celebration');
    expect(container).toHaveAttribute('data-visible', 'false');
    expect(screen.queryByText('TONK!')).not.toBeInTheDocument();
  });

  it('keeps banner visible when dismissAfterMs=0', () => {
    render(<TonkOnDealCelebration show={true} dismissAfterMs={0} winnerCards={[]} />);
    vi.advanceTimersByTime(10_000);
    expect(screen.getByTestId('tonk-on-deal-celebration')).toHaveAttribute('data-visible', 'true');
  });

  it('uses assertive aria-live so screen readers announce the win', () => {
    render(<TonkOnDealCelebration show={true} />);
    expect(screen.getByTestId('tonk-on-deal-celebration')).toHaveAttribute('aria-live', 'assertive');
  });
});
