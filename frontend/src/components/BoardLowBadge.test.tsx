import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { BoardLowBadge } from './BoardLowBadge';

const t = (key: string, opts?: Record<string, unknown>) => {
  if (key === 'boardLow.ariaPossible') return `Need ${opts?.needed} ranks`;
  return key;
};
const card = (value: number): Card => ({ design: 'SPADE', value });

describe('BoardLowBadge', () => {
  it('announces only status changes in a persistent polite live region', () => {
    const { rerender } = render(<BoardLowBadge communityCards={[card(2), card(4)]} t={t} />);
    const live = screen.getByTestId('board-low-announcement');
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toBeEmptyDOMElement();

    rerender(<BoardLowBadge communityCards={[card(2), card(4), card(11)]} t={t} />);
    expect(live).toBeEmptyDOMElement();

    rerender(<BoardLowBadge communityCards={[card(2), card(4), card(11)]} t={t} />);
    expect(live).toBeEmptyDOMElement();

    rerender(<BoardLowBadge communityCards={[card(2), card(4), card(10), card(11)]} t={t} />);
    expect(live).toBeEmptyDOMElement();

    rerender(<BoardLowBadge communityCards={[card(2), card(4), card(6), card(11)]} t={t} />);
    expect(live).toHaveTextContent('boardLow.ariaLive');

    rerender(<BoardLowBadge communityCards={[card(2), card(4), card(6), card(10), card(11)]} t={t} />);
    expect(live).toHaveTextContent('boardLow.ariaLive');
  });

  it('renders an empty live region on first render without announcing the initial status', () => {
    render(<BoardLowBadge communityCards={[card(2), card(4)]} t={t} />);
    expect(screen.getByTestId('board-low-announcement')).toHaveAttribute('aria-live', 'polite');
    expect(screen.getByTestId('board-low-announcement')).toBeEmptyDOMElement();
  });
});
