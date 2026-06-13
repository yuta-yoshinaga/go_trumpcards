import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { SuitProgressBadge } from './SuitProgressBadge';

describe('SuitProgressBadge', () => {
  it('renders four suit glyphs', () => {
    render(<SuitProgressBadge completed={0} />);
    expect(screen.getAllByTestId('suit-todo')).toHaveLength(4);
    expect(screen.queryAllByTestId('suit-done')).toHaveLength(0);
  });

  it('fills the first N glyphs as completed', () => {
    render(<SuitProgressBadge completed={2} />);
    const done = screen.getAllByTestId('suit-done');
    const todo = screen.getAllByTestId('suit-todo');
    expect(done).toHaveLength(2);
    expect(todo).toHaveLength(2);
    expect(done[0]).toHaveClass('text-ds-success');
    expect(todo[0]).toHaveClass('text-ds-text-muted');
  });

  it('fills every glyph when all suits are complete', () => {
    render(<SuitProgressBadge completed={4} />);
    expect(screen.getAllByTestId('suit-done')).toHaveLength(4);
    expect(screen.getByTestId('suit-progress')).toHaveAttribute('aria-label', '4/4');
  });

  it('clamps out-of-range counts to [0, 4]', () => {
    const { rerender } = render(<SuitProgressBadge completed={9} />);
    expect(screen.getAllByTestId('suit-done')).toHaveLength(4);
    rerender(<SuitProgressBadge completed={-3} />);
    expect(screen.getAllByTestId('suit-todo')).toHaveLength(4);
  });

  it('renders an optional label', () => {
    render(<SuitProgressBadge completed={1} label="完成" />);
    expect(screen.getByText('完成:')).toBeInTheDocument();
  });
});
