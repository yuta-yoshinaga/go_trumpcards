import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { GamePageShell } from './GamePageShell';

vi.mock('./tutorial/TutorialButton', () => ({
  TutorialButton: () => <button type="button">チュートリアル</button>,
}));

vi.mock('./ManualButton', () => ({
  ManualButton: ({ gamePath }: { gamePath: string }) => (
    <button type="button" aria-label={`manual-${gamePath}`}>
      📖
    </button>
  ),
}));

vi.mock('./motion/WinCelebration', () => ({
  WinCelebration: ({ show }: { show: boolean }) => (show ? <div data-testid="win-celebration">Win!</div> : null),
}));

vi.mock('./GameResetDialog', () => ({
  GameResetDialog: ({
    confirmOpen,
    confirmReset,
    cancelReset,
  }: {
    confirmOpen: boolean;
    confirmReset: () => void;
    cancelReset: () => void;
  }) =>
    confirmOpen ? (
      <div role="alertdialog">
        <button type="button" onClick={confirmReset}>
          確認
        </button>
        <button type="button" onClick={cancelReset}>
          キャンセル
        </button>
      </div>
    ) : null,
}));

const baseProps = {
  title: 'Hearts',
  gameThemeBg: 'bg-game-bg-blue',
  phaseName: 'Play Phase',
  isHumanTurn: false,
  gamePath: '/hearts',
  gameEndFlag: false,
  loading: false,
  confirmOpen: false,
  confirmReset: vi.fn(),
  cancelReset: vi.fn(),
};

describe('GamePageShell', () => {
  it('renders children inside the shell', () => {
    render(
      <GamePageShell {...baseProps}>
        <div data-testid="game-content">Game Content</div>
      </GamePageShell>,
    );
    expect(screen.getByTestId('game-content')).toBeInTheDocument();
    expect(screen.getByText('Game Content')).toBeInTheDocument();
  });

  it('renders a visually-hidden h1 with the title', () => {
    render(
      <GamePageShell {...baseProps}>
        <div />
      </GamePageShell>,
    );
    expect(screen.getByRole('heading', { name: 'Hearts', level: 1 })).toBeInTheDocument();
  });

  it('renders the PhaseIndicator with the phase name', () => {
    render(
      <GamePageShell {...baseProps}>
        <div />
      </GamePageShell>,
    );
    expect(screen.getByText('Play Phase')).toBeInTheDocument();
  });

  it('renders TutorialButton and ManualButton inside PhaseIndicator', () => {
    render(
      <GamePageShell {...baseProps}>
        <div />
      </GamePageShell>,
    );
    expect(screen.getByText('チュートリアル')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'manual-/hearts' })).toBeInTheDocument();
  });

  it('does not show WinCelebration when gameEndFlag is false', () => {
    render(
      <GamePageShell {...baseProps} gameEndFlag={false}>
        <div />
      </GamePageShell>,
    );
    expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
  });

  it('shows WinCelebration when gameEndFlag is true', () => {
    render(
      <GamePageShell {...baseProps} gameEndFlag={true}>
        <div />
      </GamePageShell>,
    );
    expect(screen.getByTestId('win-celebration')).toBeInTheDocument();
  });

  it('does not show GameResetDialog when confirmOpen is false', () => {
    render(
      <GamePageShell {...baseProps} confirmOpen={false}>
        <div />
      </GamePageShell>,
    );
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('shows GameResetDialog when confirmOpen is true', () => {
    render(
      <GamePageShell {...baseProps} confirmOpen={true}>
        <div />
      </GamePageShell>,
    );
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('calls confirmReset when the confirm button is clicked', () => {
    const confirmReset = vi.fn();
    render(
      <GamePageShell {...baseProps} confirmOpen={true} confirmReset={confirmReset}>
        <div />
      </GamePageShell>,
    );
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    expect(confirmReset).toHaveBeenCalledOnce();
  });

  it('calls cancelReset when the cancel button is clicked', () => {
    const cancelReset = vi.fn();
    render(
      <GamePageShell {...baseProps} confirmOpen={true} cancelReset={cancelReset}>
        <div />
      </GamePageShell>,
    );
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(cancelReset).toHaveBeenCalledOnce();
  });

  it('applies the gameThemeBg class to the outer container', () => {
    const { container } = render(
      <GamePageShell {...baseProps} gameThemeBg="bg-game-bg-red">
        <div />
      </GamePageShell>,
    );
    expect(container.firstChild).toHaveClass('bg-game-bg-red');
  });

  it('passes gamePath to ManualButton', () => {
    render(
      <GamePageShell {...baseProps} gamePath="/spades">
        <div />
      </GamePageShell>,
    );
    expect(screen.getByRole('button', { name: 'manual-/spades' })).toBeInTheDocument();
  });

  it('does not place aria-live on the outer container', () => {
    const { container } = render(
      <GamePageShell {...baseProps}>
        <div />
      </GamePageShell>,
    );
    expect(container.firstChild).not.toHaveAttribute('aria-live');
  });

  it('still forwards loading to aria-busy on the outer container', () => {
    const { container } = render(
      <GamePageShell {...baseProps} loading={true}>
        <div />
      </GamePageShell>,
    );
    expect(container.firstChild).toHaveAttribute('aria-busy', 'true');
  });
});
