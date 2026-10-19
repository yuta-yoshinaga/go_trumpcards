import { fireEvent, screen, waitFor } from '@testing-library/react';
import { useState } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi } from '../../api/gameApi';
import { renderWithProviders } from '../../test/renderWithProviders';
import { createGamePage, type GameDefinition, type GameHookBase, type GamePageContext } from './createGamePage';

vi.mock('../../api/gameApi', () => ({ actionLogApi: { acesup: vi.fn() } }));

const mockPlaySound = vi.fn();
const mockSoundValue = {
  playSound: mockPlaySound,
  muted: false,
  toggleMute: vi.fn(),
  claimExecSound: vi.fn(),
  consumeExecClaim: () => false,
};

vi.mock('../../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

interface State {
  ended: boolean;
}
interface Game extends GameHookBase<State, unknown[]> {
  exec: (...args: unknown[]) => Promise<void>;
}
const reset = vi.fn();
const giveUp = vi.fn();
const shortcut = vi.fn();
const retry = vi.fn();

const defaultBaseGame: Game = {
  state: { ended: false },
  loading: false,
  error: null,
  hintError: null,
  retry,
  handleReset: reset,
  exec: async () => undefined,
};

let currentGame: Game = { ...defaultBaseGame };
function useFakeGame() {
  return currentGame;
}
const Board = ({ state }: GamePageContext<State, Game>) => <div data-testid="board">board:{String(state.ended)}</div>;
const LiveRegion = () => <div data-testid="live">live</div>;
const FooterActions = ({ confirmGiveUpAction }: GamePageContext<State, Game>) => (
  <button type="button" onClick={confirmGiveUpAction}>
    Give up
  </button>
);

function makeDefinition(options: { customSettings?: boolean; includeGiveUp?: boolean; shortcuts?: boolean } = {}) {
  const base: Game = {
    ...defaultBaseGame,
    ...(options.includeGiveUp ? { handleGiveUp: giveUp } : {}),
  };
  currentGame = base;
  const definition: GameDefinition<State, Game, unknown[]> = {
    key: 'acesup',
    useGame: useFakeGame,
    tutorial: [],
    skeleton: { kind: 'centered', rows: [1] },
    cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
    phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
    ...(options.shortcuts
      ? { shortcuts: () => [{ key: 'x', action: shortcut, label: 'draw' }], shortcutsEnabled: () => true }
      : {}),
    ...(options.customSettings
      ? { settings: () => [{ items: [{ type: 'checkbox', id: 'custom-setting', label: 'Custom', checked: true }] }] }
      : {}),
    LiveRegion,
    Board,
    FooterActions,
  };
  return createGamePage(definition);
}

beforeEach(() => {
  vi.clearAllMocks();
  mockPlaySound.mockClear();
  localStorage.removeItem('cli-mode-acesup');
  currentGame = { ...defaultBaseGame };
});

describe('createGamePage', () => {
  it('sets inner component displayName to {key}PageContent', () => {
    const Page = makeDefinition();
    expect(Page.displayName).toBe('withTutorial(acesupPageContent)');
  });

  it('renders the configured skeleton when state is null', () => {
    const Page = makeDefinition();
    currentGame = { state: null, loading: true, error: null, retry, handleReset: reset, exec: async () => undefined };
    renderWithProviders(<Page />);
    expect(document.querySelector('[data-skeleton-section]')).toBeInTheDocument();
    expect(screen.queryByTestId('board')).not.toBeInTheDocument();
  });

  it('keeps the live region and switches the board for the CLI terminal', async () => {
    const Page = makeDefinition();
    renderWithProviders(<Page />);
    expect(screen.getByTestId('live')).toBeInTheDocument();
    expect(screen.getByTestId('board')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'CLIモードに切り替え' }));
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    expect(screen.getByTestId('live')).toBeInTheDocument();
    expect(screen.queryByTestId('board')).not.toBeInTheDocument();
  });

  it('fires shortcut keys when not loading and suppresses them when loading', () => {
    const Page = makeDefinition({ shortcuts: true });
    const view = renderWithProviders(<Page />);
    fireEvent.keyDown(document, { key: 'x' });
    expect(shortcut).toHaveBeenCalledTimes(1);
    currentGame = { ...currentGame, loading: true };
    view.rerender(<Page />);
    fireEvent.keyDown(document, { key: 'x' });
    expect(shortcut).toHaveBeenCalledTimes(1);
  });

  it('fires shortcut keys when shortcutsEnabled is omitted (defaults to true)', () => {
    const Page = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      shortcuts: () => [{ key: 'x', action: shortcut }],
      Board,
    });
    renderWithProviders(<Page />);
    fireEvent.keyDown(document, { key: 'x' });
    expect(shortcut).toHaveBeenCalledTimes(1);
  });

  it('suppresses shortcuts when the definition disables them', () => {
    const DisabledPage = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      shortcuts: () => [{ key: 'x', action: shortcut }],
      shortcutsEnabled: () => false,
      Board,
      LiveRegion,
    });
    renderWithProviders(<DisabledPage />);
    fireEvent.keyDown(document, { key: 'x' });
    expect(shortcut).not.toHaveBeenCalled();
  });

  it('suppresses shortcuts and disables reset button when def.busy returns true while loading is false', () => {
    const Page = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      busy: () => true,
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      shortcuts: () => [{ key: 'x', action: shortcut }],
      Board,
    });
    currentGame = { ...currentGame, loading: false };
    renderWithProviders(<Page />);
    const resetButton = screen.getByRole('button', { name: 'リセット' });
    expect(resetButton).toBeDisabled();

    fireEvent.keyDown(document, { key: 'x' });
    expect(shortcut).not.toHaveBeenCalled();
  });

  it('uses default hint settings or custom settings', () => {
    const DefaultPage = makeDefinition();
    const view = renderWithProviders(<DefaultPage />);
    fireEvent.click(screen.getByText('設定'));
    expect(screen.getByLabelText('ヒント表示')).toBeInTheDocument();
    view.unmount();
    const CustomPage = makeDefinition({ customSettings: true });
    renderWithProviders(<CustomPage />);
    fireEvent.click(screen.getByText('設定'));
    expect(screen.getByLabelText('Custom')).toBeInTheDocument();
  });

  it('give-up positive case: clicking Give up then confirming calls handleGiveUp once', () => {
    const Page = makeDefinition({ includeGiveUp: true });
    renderWithProviders(<Page />);
    fireEvent.click(screen.getByRole('button', { name: 'Give up' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    expect(giveUp).toHaveBeenCalledTimes(1);
  });

  it('give-up absent case: no give-up dialog props reach the shell and confirmGiveUpAction is undefined', () => {
    let capturedAction: (() => void) | undefined = () => undefined;
    const SlotWithGiveUp = ({ confirmGiveUpAction }: GamePageContext<State, Game>) => {
      capturedAction = confirmGiveUpAction;
      return <div data-testid="giveup-slot" />;
    };
    const PageWithoutGiveUp = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      Board: SlotWithGiveUp,
    });
    renderWithProviders(<PageWithoutGiveUp />);
    expect(screen.getByTestId('giveup-slot')).toBeInTheDocument();
    expect(capturedAction).toBeUndefined();
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('groups game footer actions and reset together', () => {
    const Page = makeDefinition();
    renderWithProviders(<Page />);
    const giveUpButton = screen.getByRole('button', { name: 'Give up' });
    const resetButton = screen.getByRole('button', { name: 'リセット' });
    expect(giveUpButton.parentElement).toHaveClass('flex', 'gap-2', 'items-center', 'flex-wrap');
    expect(resetButton.parentElement).toBe(giveUpButton.parentElement);
  });

  it('hides the action log before resetting', () => {
    const Page = makeDefinition();
    currentGame = { ...currentGame, state: { ended: true } };
    vi.mocked(actionLogApi.acesup).mockResolvedValue({ entries: [] });
    renderWithProviders(<Page />);
    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    return waitFor(() => expect(screen.getByRole('button', { name: '閉じる' })).toBeInTheDocument()).then(() => {
      fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
      expect(reset).toHaveBeenCalledTimes(1);
      expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument();
    });
  });

  it('shows error instead of hintError when both are set, and retry button calls retry', () => {
    const Page = makeDefinition();
    currentGame = { ...currentGame, error: 'Game failure', hintError: 'Hint failure' };
    renderWithProviders(<Page />);
    const alert = screen.getByRole('alert');
    expect(alert).toHaveTextContent('Game failure');
    expect(alert).not.toHaveTextContent('Hint failure');

    fireEvent.click(screen.getByRole('button', { name: '再試行' }));
    expect(retry).toHaveBeenCalledTimes(1);
  });

  it('shows hintError when the game error is null', () => {
    const Page = makeDefinition();
    currentGame = { ...currentGame, hintError: 'Hint unavailable' };
    renderWithProviders(<Page />);
    expect(screen.getByRole('alert')).toHaveTextContent('Hint unavailable');
  });

  it('useView result reaches ctx.view and persists across a CLI toggle round-trip', async () => {
    function useTestView() {
      const [count, setCount] = useState(0);
      return { count, increment: () => setCount((c) => c + 1) };
    }
    type TestView = ReturnType<typeof useTestView>;
    const ViewBoard = ({ view }: GamePageContext<State, Game, TestView>) => (
      <div data-testid="board">
        <span data-testid="view-count">{view.count}</span>
        <button type="button" onClick={view.increment}>
          Increment
        </button>
      </div>
    );

    const PageWithView = createGamePage<State, Game, unknown[], TestView>({
      key: 'acesup',
      useGame: useFakeGame,
      useView: useTestView,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      Board: ViewBoard,
    });

    renderWithProviders(<PageWithView />);
    expect(screen.getByTestId('view-count')).toHaveTextContent('0');

    fireEvent.click(screen.getByRole('button', { name: 'Increment' }));
    expect(screen.getByTestId('view-count')).toHaveTextContent('1');

    // Toggle to CLI mode - Board unmounts
    fireEvent.click(screen.getByRole('button', { name: 'CLIモードに切り替え' }));
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    expect(screen.queryByTestId('view-count')).not.toBeInTheDocument();

    // Toggle back to GUI mode - Board remounts, count must still be 1
    fireEvent.click(screen.getByRole('button', { name: 'GUIモードに切り替え' }));
    await waitFor(() => expect(screen.getByTestId('board')).toBeInTheDocument());
    expect(screen.getByTestId('view-count')).toHaveTextContent('1');
  });

  it('cli as a function: changing state changes the config used, and onCommandSuccess is called', async () => {
    const onCommandSuccess = vi.fn();
    interface CliState {
      ended: boolean;
      counter: number;
    }
    let setCliState: React.Dispatch<React.SetStateAction<CliState>> | undefined;
    const mockExec = vi.fn(async (_cmd: string) => {
      setCliState?.({ ended: false, counter: 42 });
    });
    function useCliTestGame() {
      const [state, setState] = useState<CliState>({ ended: false, counter: 1 });
      setCliState = setState;
      return {
        state,
        loading: false,
        error: null,
        hintError: null,
        retry,
        handleReset: reset,
        exec: mockExec,
      };
    }
    const CliPage = createGamePage<CliState, ReturnType<typeof useCliTestGame>, [string]>({
      key: 'acesup',
      useGame: useCliTestGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: ({ state }) => ({
        parseCommand: (cmd) => ({ args: [cmd] }),
        formatResponse: () => `counter:${state?.counter}`,
        helpText: ['help'],
        onCommandSuccess,
      }),
      phase: () => ({ name: 'playing', ended: false }),
      Board: () => <div data-testid="board">board</div>,
    });
    renderWithProviders(<CliPage />);
    fireEvent.click(screen.getByRole('button', { name: 'CLIモードに切り替え' }));
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());

    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'draw' } });
    fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    expect(onCommandSuccess).toHaveBeenCalledWith(['draw']);

    await waitFor(() => expect(screen.getAllByText('counter:42').length).toBeGreaterThanOrEqual(1));
  });

  it('passes shell props (isHumanTurn, lossShow, headerEnd) to GamePageShell', () => {
    const Page = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: () => ({ name: 'ended', ended: true }),
      shell: () => ({
        isHumanTurn: true,
        lossShow: true,
        headerEnd: <span data-testid="header-end">End</span>,
      }),
      Board,
    });
    renderWithProviders(<Page />);
    expect(screen.getByText('あなたのターン')).toBeInTheDocument();
    expect(screen.getByTestId('header-end')).toBeInTheDocument();
    expect(mockPlaySound).toHaveBeenCalledWith('lossThud');
  });

  it('calls custom reset instead of handleReset when reset is provided', () => {
    const customReset = vi.fn();
    const Page = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      reset: customReset,
      Board,
    });
    renderWithProviders(<Page />);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    expect(customReset).toHaveBeenCalledTimes(1);
    expect(reset).not.toHaveBeenCalled();
  });

  it('applies custom contentClassName or defaults to standard padding', () => {
    const CustomPage = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      contentClassName: 'custom-pad-123',
      Board,
    });
    const { unmount } = renderWithProviders(<CustomPage />);
    const customContainer = screen.getByTestId('board').parentElement;
    expect(customContainer).toHaveClass('custom-pad-123');
    expect(customContainer).not.toHaveClass('pt-3', 'px-4', 'lg:px-8');
    unmount();

    const DefaultPage = makeDefinition();
    renderWithProviders(<DefaultPage />);
    const defaultContainer = screen.getByTestId('board').parentElement;
    expect(defaultContainer).toHaveClass('pt-3', 'px-4', 'lg:px-8');
  });

  it('renders HeaderExtra, resetTutorialId, and shortcutsTestId attributes', () => {
    const Page = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      shortcuts: () => [{ key: 'x', action: vi.fn(), label: 'draw' }],
      resetTutorialId: 'tut-reset-btn',
      shortcutsTestId: 'test-shortcuts-panel',
      HeaderExtra: () => <span data-testid="header-extra">extra-info</span>,
      Board,
    });
    renderWithProviders(<Page />);
    expect(screen.getByTestId('header-extra')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'リセット' })).toHaveAttribute('data-tutorial', 'tut-reset-btn');
    expect(screen.getByTestId('test-shortcuts-panel')).toBeInTheDocument();
  });

  it('renders LandscapeBanner in GUI mode only and unmounts it in CLI mode', async () => {
    const Page = makeDefinition();
    renderWithProviders(<Page />);
    expect(screen.getByText('横画面での表示を推奨します')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'CLIモードに切り替え' }));
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    expect(screen.queryByText('横画面での表示を推奨します')).not.toBeInTheDocument();
  });
});
