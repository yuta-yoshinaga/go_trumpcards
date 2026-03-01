import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { StatusBadge } from './StatusBadge';

describe('StatusBadge', () => {
  it('renders success variant with green background and white text', () => {
    render(<StatusBadge variant="success">上がり</StatusBadge>);
    const badge = screen.getByText('上がり');
    expect(badge).toBeInTheDocument();
    expect(badge.style.background).toBe('rgb(92, 184, 92)');
    expect(badge.style.color).toBe('rgb(255, 255, 255)');
  });

  it('renders warning variant with orange background and bold dark text', () => {
    render(<StatusBadge variant="warning">考え中...</StatusBadge>);
    const badge = screen.getByText('考え中...');
    expect(badge).toBeInTheDocument();
    expect(badge.style.background).toBe('rgb(240, 173, 78)');
    expect(badge.style.color).toBe('rgb(34, 34, 34)');
    expect(badge.style.fontWeight).toBe('bold');
  });
});
