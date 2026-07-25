import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { SoundProvider } from '../providers/SoundProvider';
import { GamePageShell } from './GamePageShell';

// Track plays per sound file so the central-tap tests can assert WHICH
// sound fired (the global setup.ts mock can't distinguish Howl instances).
const { playCalls } = vi.hoisted(() => ({ playCalls: [] as string[] }));
vi.mock('howler', () => ({
  Howl: class MockHowl {
    private src: string;
    constructor(opts: { src: string[] }) {
      this.src = opts.src[0];
    }
    play() {
      playCalls.push(this.src);
      return 1;
    }
    volume() {}
    rate() {}
  },
  Howler: { ctx: { state: 'running' } },
}));

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
  WinCelebration: ({ show, onCelebrate }: { show: boolean; onCelebrate?: () => void }) =>
    show ? (
      <button type="button" data-testid="win-celebration" onClick={() => onCelebrate?.()}>
        Win!
      </button>
    ) : null,
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

const propsWithoutTurn = {
  title: 'TriPeaks',
  gameThemeBg: 'bg-game-bg-blue',
  phaseName: 'Play Phase',
  gamePath: '/tripeaks',
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

  it('uses winShow override when provided (suppress celebration even at game end)', () => {
    render(
      <GamePageShell {...baseProps} gameEndFlag={true} winShow={false}>
        <div />
      </GamePageShell>,
    );
    expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
  });

  it('uses winShow override when provided (show celebration before gameEndFlag flips)', () => {
    render(
      <GamePageShell {...baseProps} gameEndFlag={false} winShow={true}>
        <div />
      </GamePageShell>,
    );
    expect(screen.getByTestId('win-celebration')).toBeInTheDocument();
  });

  it('forwards onCelebrate to WinCelebration', () => {
    const onCelebrate = vi.fn();
    render(
      <GamePageShell {...baseProps} gameEndFlag={true} onCelebrate={onCelebrate}>
        <div />
      </GamePageShell>,
    );
    fireEvent.click(screen.getByTestId('win-celebration'));
    expect(onCelebrate).toHaveBeenCalledOnce();
  });

  it('renders headerEnd after ManualButton (right of buttons)', () => {
    render(
      <GamePageShell
        {...baseProps}
        headerExtra={<span data-testid="header-extra">extra</span>}
        headerEnd={<span data-testid="header-end">end</span>}
      >
        <div />
      </GamePageShell>,
    );
    const extra = screen.getByTestId('header-extra');
    const manual = screen.getByRole('button', { name: 'manual-/hearts' });
    const end = screen.getByTestId('header-end');
    // headerExtra → TutorialButton → ManualButton → headerEnd, in DOM order.
    expect(extra.compareDocumentPosition(manual) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(manual.compareDocumentPosition(end) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it('remounts the outer container when outerKey changes (drives shake animation)', () => {
    const { container, rerender } = render(
      <GamePageShell {...baseProps} outerKey={0}>
        <div data-testid="game-body" />
      </GamePageShell>,
    );
    const firstBody = screen.getByTestId('game-body');
    rerender(
      <GamePageShell {...baseProps} outerKey={1}>
        <div data-testid="game-body" />
      </GamePageShell>,
    );
    // Changing outerKey forces React to unmount the previous subtree and mount a new one,
    // which is what restarts the shake CSS animation. Verify by ensuring the body element
    // is no longer the same DOM node.
    expect(container.contains(firstBody)).toBe(false);
    expect(screen.getByTestId('game-body')).toBeInTheDocument();
  });

  it('omits the turn-indicator span when isHumanTurn is not provided', () => {
    render(
      <GamePageShell {...propsWithoutTurn}>
        <div />
      </GamePageShell>,
    );
    // PhaseIndicator should still render the phase name…
    expect(screen.getByText('Play Phase')).toBeInTheDocument();
    // …but no "your turn / waiting" label is rendered when the prop is undefined.
    expect(screen.queryByText('あなたのターン')).not.toBeInTheDocument();
    expect(screen.queryByText('待機中')).not.toBeInTheDocument();
  });

  it('establishes a positioning context so absolute-positioned children stay inside the page (#1900)', () => {
    // Regression: ISSUE-005 — without `relative` on the outer container,
    // absolutely-positioned descendants like CpuActionToast escape to the
    // viewport and overlap the mobile NavBar's brand bar. Found by /qa on
    // 2026-05-20 (mobile poker, top-left "Trump Cards" overlap).
    const { container } = render(
      <GamePageShell {...baseProps}>
        <div data-testid="anchor" />
      </GamePageShell>,
    );
    const outer = container.firstElementChild as HTMLElement;
    expect(outer.className).toContain('relative');
  });

  describe('central sound taps', () => {
    beforeEach(() => {
      playCalls.length = 0;
      vi.useFakeTimers();
    });
    afterEach(() => {
      vi.useRealTimers();
      localStorage.clear();
    });

    type TapProps = Omit<typeof baseProps, 'isHumanTurn'> & {
      isHumanTurn?: boolean;
      winShow?: boolean;
      onCelebrate?: () => void;
    };

    function renderWithSound(props: TapProps) {
      return render(
        <SoundProvider>
          <GamePageShell {...props}>
            <div />
          </GamePageShell>
        </SoundProvider>,
      );
    }

    it('plays winFanfare when the celebration fires', () => {
      renderWithSound({ ...baseProps, gameEndFlag: true });
      fireEvent.click(screen.getByTestId('win-celebration'));
      expect(playCalls).toContain('/sounds/win-fanfare.ogg');
      expect(playCalls).not.toContain('/sounds/loss-thud.ogg');
    });

    it('still invokes the page onCelebrate callback after the central fanfare', () => {
      const onCelebrate = vi.fn();
      renderWithSound({ ...baseProps, gameEndFlag: true, onCelebrate });
      fireEvent.click(screen.getByTestId('win-celebration'));
      expect(onCelebrate).toHaveBeenCalledTimes(1);
      expect(playCalls).toContain('/sounds/win-fanfare.ogg');
    });

    it('plays lossThud when the game ends without a win (winShow === false)', () => {
      renderWithSound({ ...baseProps, gameEndFlag: true, winShow: false });
      expect(playCalls).toContain('/sounds/loss-thud.ogg');
      expect(playCalls).not.toContain('/sounds/win-fanfare.ogg');
    });

    it('never plays lossThud when winShow is undefined (celebration-mirror pages)', () => {
      renderWithSound({ ...baseProps, gameEndFlag: true });
      expect(playCalls).not.toContain('/sounds/loss-thud.ogg');
    });

    it('resets the loss latch on a new game (gameEndFlag falls, then rises)', () => {
      const { rerender } = renderWithSound({ ...baseProps, gameEndFlag: true, winShow: false });
      expect(playCalls.filter((p) => p === '/sounds/loss-thud.ogg')).toHaveLength(1);

      // Clear the provider's 3s dedupe window, then start a new round.
      vi.advanceTimersByTime(3100);
      rerender(
        <SoundProvider>
          <GamePageShell {...baseProps} gameEndFlag={false} winShow={false}>
            <div />
          </GamePageShell>
        </SoundProvider>,
      );
      rerender(
        <SoundProvider>
          <GamePageShell {...baseProps} gameEndFlag={true} winShow={false}>
            <div />
          </GamePageShell>
        </SoundProvider>,
      );
      expect(playCalls.filter((p) => p === '/sounds/loss-thud.ogg')).toHaveLength(2);
    });

    it('plays turnTick exactly once on the false→true edge of isHumanTurn', () => {
      const { rerender } = renderWithSound({ ...baseProps, isHumanTurn: false });
      expect(playCalls).not.toContain('/sounds/turn-tick.ogg');

      rerender(
        <SoundProvider>
          <GamePageShell {...baseProps} isHumanTurn={true}>
            <div />
          </GamePageShell>
        </SoundProvider>,
      );
      expect(playCalls.filter((p) => p === '/sounds/turn-tick.ogg')).toHaveLength(1);

      // Staying on the human's turn must not re-tick.
      rerender(
        <SoundProvider>
          <GamePageShell {...baseProps} isHumanTurn={true} phaseName="Other">
            <div />
          </GamePageShell>
        </SoundProvider>,
      );
      expect(playCalls.filter((p) => p === '/sounds/turn-tick.ogg')).toHaveLength(1);
    });

    it('never ticks when isHumanTurn is omitted (solitaire pages)', () => {
      const { rerender } = renderWithSound({ ...propsWithoutTurn });
      rerender(
        <SoundProvider>
          <GamePageShell {...propsWithoutTurn} phaseName="Other">
            <div />
          </GamePageShell>
        </SoundProvider>,
      );
      expect(playCalls).not.toContain('/sounds/turn-tick.ogg');
    });
  });
});
