import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { GameResetButton } from './GameResetButton';

describe('GameResetButton', () => {
  it('shows "リセット" and routes click through requestConfirm mid-game', () => {
    const onReset = vi.fn();
    const requestConfirm = vi.fn();
    render(<GameResetButton isGameEnd={false} onReset={onReset} requestConfirm={requestConfirm} />);

    const button = screen.getByRole('button', { name: 'リセット' });
    fireEvent.click(button);

    expect(requestConfirm).toHaveBeenCalledTimes(1);
    expect(onReset).not.toHaveBeenCalled();
    // The callback passed to requestConfirm invokes onReset.
    requestConfirm.mock.calls[0][0]();
    expect(onReset).toHaveBeenCalledTimes(1);
  });

  it('shows "次のゲーム" and invokes onReset directly at game end', () => {
    const onReset = vi.fn();
    const requestConfirm = vi.fn();
    render(<GameResetButton isGameEnd={true} onReset={onReset} requestConfirm={requestConfirm} />);

    const button = screen.getByRole('button', { name: '次のゲーム' });
    fireEvent.click(button);

    expect(onReset).toHaveBeenCalledTimes(1);
    expect(requestConfirm).not.toHaveBeenCalled();
  });

  it('disables the button when loading', () => {
    render(
      <GameResetButton isGameEnd={false} onReset={() => undefined} requestConfirm={() => undefined} loading={true} />,
    );
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('forwards dataTutorial attribute', () => {
    render(
      <GameResetButton
        isGameEnd={false}
        onReset={() => undefined}
        requestConfirm={() => undefined}
        dataTutorial="hearts-reset"
      />,
    );
    expect(screen.getByRole('button')).toHaveAttribute('data-tutorial', 'hearts-reset');
  });
});
