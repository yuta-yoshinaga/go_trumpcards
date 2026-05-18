/**
 * @vitest-environment jsdom
 */
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { TOTAL_QUESTIONS } from '../../constants/discoverAxes';
import { DiscoverSkeleton } from './DiscoverSkeleton';

describe('DiscoverSkeleton', () => {
  it('renders a status region with aria-busy', () => {
    render(<DiscoverSkeleton />);
    const status = screen.getByRole('status');
    expect(status).toHaveAttribute('aria-busy', 'true');
  });

  it('renders TOTAL_QUESTIONS card placeholders + 4 option rows', () => {
    const { container } = render(<DiscoverSkeleton />);
    // 8 deck card placeholders are <li> elements with aria-hidden parent.
    const deck = container.querySelector('ul[aria-hidden="true"]');
    expect(deck?.children).toHaveLength(TOTAL_QUESTIONS);
    // 4 option-row placeholders (separate <ul>).
    const optionRows = container.querySelectorAll('ul')[1]?.children;
    expect(optionRows).toHaveLength(4);
  });
});
