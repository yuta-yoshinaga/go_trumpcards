import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { SuitProgressBadge } from './SuitProgressBadge';

describe('SuitProgressBadge', () => {
  it('renders four suit glyphs', () => {
    render(<SuitProgressBadge completedMask={0} />);
    expect(screen.getAllByTestId('suit-todo')).toHaveLength(4);
    expect(screen.queryAllByTestId('suit-done')).toHaveLength(0);
  });

  it('fills only heart and diamond for mask 12', () => {
    render(<SuitProgressBadge completedMask={12} />);
    const done = screen.getAllByTestId('suit-done');
    const todo = screen.getAllByTestId('suit-todo');
    expect(done).toHaveLength(2);
    expect(todo).toHaveLength(2);
    expect(screen.getAllByTestId('suit-todo')[0]).toHaveTextContent('♠');
    expect(screen.getAllByTestId('suit-todo')[1]).toHaveTextContent('♣');
    expect(screen.getAllByTestId('suit-done')[0]).toHaveTextContent('♥');
    expect(screen.getAllByTestId('suit-done')[1]).toHaveTextContent('♦');
    expect(screen.getByTestId('suit-progress')).toHaveAttribute('aria-label', '2/4');
  });

  it('fills every glyph when all suits are complete', () => {
    render(<SuitProgressBadge completedMask={15} />);
    expect(screen.getAllByTestId('suit-done')).toHaveLength(4);
    expect(screen.getByTestId('suit-progress')).toHaveAttribute('aria-label', '4/4');
  });

  it('renders all suits as todo for mask 0', () => {
    render(<SuitProgressBadge completedMask={0} />);
    expect(screen.getAllByTestId('suit-todo')).toHaveLength(4);
  });

  it('renders an optional label', () => {
    render(<SuitProgressBadge completedMask={1} label="完成" />);
    expect(screen.getByText('完成:')).toBeInTheDocument();
  });
});
