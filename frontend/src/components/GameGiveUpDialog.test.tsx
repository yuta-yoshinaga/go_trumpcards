import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { GameGiveUpDialog } from './GameGiveUpDialog';

describe('GameGiveUpDialog', () => {
  it('renders nothing when closed', () => {
    render(<GameGiveUpDialog giveUpConfirmOpen={false} confirmGiveUp={vi.fn()} cancelGiveUp={vi.fn()} />);
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('renders give-up-specific copy (not reset copy) when open', () => {
    render(<GameGiveUpDialog giveUpConfirmOpen={true} confirmGiveUp={vi.fn()} cancelGiveUp={vi.fn()} />);
    // ja translations are loaded in test/setup.ts.
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    expect(screen.getByText('投了すると現在のゲームは終了します。よろしいですか？')).toBeInTheDocument();
  });

  it('fires confirmGiveUp on confirm click', () => {
    const confirmGiveUp = vi.fn();
    render(<GameGiveUpDialog giveUpConfirmOpen={true} confirmGiveUp={confirmGiveUp} cancelGiveUp={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    expect(confirmGiveUp).toHaveBeenCalledTimes(1);
  });

  it('fires cancelGiveUp on cancel click', () => {
    const cancelGiveUp = vi.fn();
    render(<GameGiveUpDialog giveUpConfirmOpen={true} confirmGiveUp={vi.fn()} cancelGiveUp={cancelGiveUp} />);
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(cancelGiveUp).toHaveBeenCalledTimes(1);
  });
});
