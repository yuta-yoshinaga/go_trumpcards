import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ManualModal } from './ManualModal';

vi.mock('../constants/manualTexts', () => ({
  manualTexts: {
    '/': '# BlackJack\n\nTest **bold** content\n\n| A | B |\n|---|---|\n| 1 | 2 |',
    '/poker': '# Poker\n\nPoker manual',
  },
}));

describe('ManualModal', () => {
  it('renders nothing when closed', () => {
    const { container } = render(<ManualModal open={false} onClose={vi.fn()} gamePath="/" />);
    expect(container.innerHTML).toBe('');
  });

  it('renders markdown content when open', () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText('BlackJack')).toBeInTheDocument();
    expect(screen.getByText('bold')).toBeInTheDocument();
  });

  it('renders GFM table', () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    expect(screen.getByRole('table')).toBeInTheDocument();
  });

  it('renders different manual for different gamePath', () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/poker" />);
    expect(screen.getByText('Poker')).toBeInTheDocument();
  });

  it('renders empty content for unknown gamePath', () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/unknown" />);
    const dialog = screen.getByRole('dialog');
    // Should still render dialog but with no markdown content
    expect(dialog).toBeInTheDocument();
  });

  it('calls onClose when close button is clicked', () => {
    const onClose = vi.fn();
    render(<ManualModal open={true} onClose={onClose} gamePath="/" />);
    fireEvent.click(screen.getByRole('button', { name: '×' }));
    expect(onClose).toHaveBeenCalledOnce();
  });

  it('calls onClose when overlay is clicked', () => {
    const onClose = vi.fn();
    render(<ManualModal open={true} onClose={onClose} gamePath="/" />);
    // Click the overlay (presentation role)
    const overlay = screen.getByRole('presentation');
    fireEvent.click(overlay);
    expect(onClose).toHaveBeenCalledOnce();
  });

  it('does not call onClose when dialog content is clicked', () => {
    const onClose = vi.fn();
    render(<ManualModal open={true} onClose={onClose} gamePath="/" />);
    fireEvent.click(screen.getByRole('dialog'));
    expect(onClose).not.toHaveBeenCalled();
  });

  it('calls onClose on Escape key', () => {
    const onClose = vi.fn();
    render(<ManualModal open={true} onClose={onClose} gamePath="/" />);
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledOnce();
  });

  it('has aria-modal and aria-label attributes', () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    const dialog = screen.getByRole('dialog');
    expect(dialog).toHaveAttribute('aria-modal', 'true');
    expect(dialog).toHaveAttribute('aria-label', 'Game Manual');
  });
});
