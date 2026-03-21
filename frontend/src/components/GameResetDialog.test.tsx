import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { GameResetDialog } from './GameResetDialog';

describe('GameResetDialog', () => {
  it('renders nothing when confirmOpen is false', () => {
    const { container } = render(<GameResetDialog confirmOpen={false} confirmReset={vi.fn()} cancelReset={vi.fn()} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders the dialog with common translation keys when open', () => {
    render(<GameResetDialog confirmOpen={true} confirmReset={vi.fn()} cancelReset={vi.fn()} />);
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
    expect(screen.getByText('リセット確認')).toBeInTheDocument();
    expect(screen.getByText('本当にゲームをリセットしますか？')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'キャンセル' })).toBeInTheDocument();
  });

  it('calls confirmReset when confirm button is clicked', () => {
    const confirmReset = vi.fn();
    render(<GameResetDialog confirmOpen={true} confirmReset={confirmReset} cancelReset={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    expect(confirmReset).toHaveBeenCalledOnce();
  });

  it('calls cancelReset when cancel button is clicked', () => {
    const cancelReset = vi.fn();
    render(<GameResetDialog confirmOpen={true} confirmReset={vi.fn()} cancelReset={cancelReset} />);
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(cancelReset).toHaveBeenCalledOnce();
  });
});
