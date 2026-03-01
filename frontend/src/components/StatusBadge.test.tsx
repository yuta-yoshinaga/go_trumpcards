import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { StatusBadge } from './StatusBadge';

describe('StatusBadge', () => {
  it('renders success variant', () => {
    const { container } = render(<StatusBadge variant="success">上がり</StatusBadge>);
    expect(screen.getByText('上がり')).toBeInTheDocument();
    expect(container.firstChild).toMatchSnapshot();
  });

  it('renders warning variant', () => {
    const { container } = render(<StatusBadge variant="warning">考え中...</StatusBadge>);
    expect(screen.getByText('考え中...')).toBeInTheDocument();
    expect(container.firstChild).toMatchSnapshot();
  });
});
