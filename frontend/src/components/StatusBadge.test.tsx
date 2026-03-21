import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { StatusBadge } from './StatusBadge';

describe('StatusBadge', () => {
  it('renders success variant', () => {
    render(<StatusBadge variant="success">上がり</StatusBadge>);
    const badge = screen.getByText('上がり');
    expect(badge).toBeInTheDocument();
    expect(badge.tagName).toBe('SPAN');
    expect(badge.className).toContain('bg-game-status-active');
  });

  it('renders warning variant', () => {
    render(<StatusBadge variant="warning">考え中...</StatusBadge>);
    const badge = screen.getByText('考え中...');
    expect(badge).toBeInTheDocument();
    expect(badge.tagName).toBe('SPAN');
    expect(badge.className).toContain('bg-game-status-waiting');
  });

  it('renders danger variant', () => {
    render(<StatusBadge variant="danger">容疑者</StatusBadge>);
    const badge = screen.getByText('容疑者');
    expect(badge).toBeInTheDocument();
    expect(badge.tagName).toBe('SPAN');
    expect(badge.className).toContain('bg-game-status-out');
  });
});
