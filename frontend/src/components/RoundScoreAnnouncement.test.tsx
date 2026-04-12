import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { RoundScoreAnnouncement } from './RoundScoreAnnouncement';

describe('RoundScoreAnnouncement', () => {
  const entries = [
    { name: 'You', roundScore: 13, cumulativeScore: 26 },
    { name: 'CPU 1', roundScore: 0, cumulativeScore: 7 },
  ];

  it('renders an sr-only live region', () => {
    render(<RoundScoreAnnouncement active={false} entries={entries} />);
    const status = screen.getByRole('status');
    expect(status).toHaveClass('sr-only');
    expect(status).toHaveAttribute('aria-live', 'polite');
    expect(status).toHaveAttribute('aria-atomic', 'true');
  });

  it('is empty when inactive', () => {
    render(<RoundScoreAnnouncement active={false} entries={entries} />);
    expect(screen.getByRole('status').textContent).toBe('');
  });

  it('announces round deltas when active', () => {
    render(<RoundScoreAnnouncement active={true} entries={entries} />);
    const text = screen.getByRole('status').textContent ?? '';
    expect(text.length).toBeGreaterThan(0);
    expect(text).toContain('You');
    expect(text).toContain('13');
    expect(text).toContain('26');
    expect(text).toContain('CPU 1');
  });

  it('handles an empty entries list', () => {
    render(<RoundScoreAnnouncement active={true} entries={[]} />);
    expect(screen.getByRole('status').textContent).not.toBe('');
  });
});
