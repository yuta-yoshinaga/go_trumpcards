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
const cardToggle = vi.fn();
const cardConfirm = vi.fn();
const cardClear = vi.fn();
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

  it('binds card navigation, disables it while busy, and renders its shortcuts instead of action shortcuts', () => {
    const Page = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      busy: (game) => game.loading,
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      cardNav: () => ({
        cardCount: 2,
        onToggle: cardToggle,
        onConfirm: cardConfirm,
        onClear: cardClear,
        enabled: true,
      }),
      shortcuts: () => [{ key: 'x', action: shortcut, label: 'draw' }],
      shortcutsTestId: 'card-shortcuts',
      Board,
    });
    const view = renderWithProviders(<Page />);
    fireEvent.keyDown(document, { key: '1' });
    fireEvent.keyDown(document, { key: 'Enter' });
    expect(cardToggle).toHaveBeenCalledWith(0);
    expect(cardConfirm).toHaveBeenCalledTimes(1);
    expect(screen.getByTestId('card-shortcuts')).toBeInTheDocument();
    const actionShortcutSummary = view.container.querySelectorAll('footer details summary')[1];
    expect(actionShortcutSummary).toBeDefined();
    if (actionShortcutSummary) fireEvent.click(actionShortcutSummary);
    expect(screen.getByText('山札をめくる')).toBeInTheDocument();

    currentGame = { ...currentGame, loading: true };
    view.rerender(<Page />);
    fireEvent.keyDown(document, { key: '1' });
    fireEvent.keyDown(document, { key: 'Enter' });
    expect(cardToggle).toHaveBeenCalledTimes(1);
    expect(cardConfirm).toHaveBeenCalledTimes(1);
  });

  it('uses action shortcuts when card navigation is absent', () => {
    const Page = makeDefinition({ shortcuts: true });
    renderWithProviders(<Page />);
    expect(screen.getByText('キーボードショートカット')).toBeInTheDocument();
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

  it('places FooterHand before errors and FooterStatus before the actions row', () => {
    const Page = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      FooterHand: () => <div data-testid="footer-hand" />,
      FooterStatus: () => <div data-testid="footer-status" />,
      actionsTutorialId: 'actions-tutorial',
      Board,
    });
    currentGame = { ...currentGame, error: 'failure' };
    renderWithProviders(<Page />);
    const hand = screen.getByTestId('footer-hand');
    const error = screen.getByRole('alert');
    const status = screen.getByTestId('footer-status');
    const actions = screen.getByRole('button', { name: 'リセット' }).parentElement;
    expect(hand.compareDocumentPosition(error) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(error.compareDocumentPosition(status) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(actions).not.toBeNull();
    if (actions) {
      expect(status.compareDocumentPosition(actions) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
      expect(actions).toHaveAttribute('data-tutorial', 'actions-tutorial');
      expect(actions.closest('footer')).not.toBeNull();
    }
    expect(hand.closest('footer')).not.toBeNull();
    expect(error.closest('footer')).not.toBeNull();
    expect(status.closest('footer')).not.toBeNull();
  });

  it('uses the pinned footer form and moves supporting content after the action log', () => {
    const Page = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      shortcuts: () => [{ key: 'x', action: shortcut, label: 'draw' }],
      shortcutsTestId: 'pinned-shortcuts',
      primaryActions: () => <button type="button">Primary</button>,
      message: () => <span>Guidance</span>,
      FooterHand: () => <div data-testid="pinned-hand" />,
      FooterStatus: () => <div data-testid="pinned-status" />,
      FooterActions: () => <div data-testid="pinned-actions" />,
      Board,
    });
    currentGame = { ...currentGame, state: { ended: true }, error: 'failure' };
    renderWithProviders(<Page />);

    const content = screen.getByTestId('game-footer-content');
    const pinnedActions = screen.getByTestId('game-footer-actions');
    const message = screen.getByTestId('game-footer-message');
    const actionLog = screen.getByRole('button', { name: '棋譜を見る' });
    const error = screen.getByRole('alert');
    const status = screen.getByTestId('pinned-status');
    const actions = screen.getByTestId('pinned-actions').parentElement;
    const shortcuts = screen.getByTestId('pinned-shortcuts');
    const scrollingContent = actionLog.closest('[class*="overflow-y-auto"]');
    expect(content).toContainElement(screen.getByTestId('pinned-hand'));
    expect(pinnedActions).toHaveTextContent('Primary');
    expect(message).toHaveTextContent('Guidance');
    expect(error.closest('[data-testid="game-footer-content"]')).toBeNull();
    expect(status.closest('[data-testid="game-footer-content"]')).toBeNull();
    expect(actions?.closest('[data-testid="game-footer-content"]')).toBeNull();
    expect(shortcuts.closest('[data-testid="game-footer-content"]')).toBeNull();
    expect(scrollingContent).toContainElement(error);
    expect(scrollingContent).toContainElement(status);
    if (actions) expect(scrollingContent).toContainElement(actions);
    expect(scrollingContent).toContainElement(shortcuts);
    expect(error.closest('footer')).toBeNull();
    expect(status.closest('footer')).toBeNull();
    if (actions) expect(actions.closest('footer')).toBeNull();
    expect(shortcuts.closest('footer')).toBeNull();
    expect(actionLog.compareDocumentPosition(error) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(error.compareDocumentPosition(status) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    if (actions) expect(status.compareDocumentPosition(actions) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    if (actions) expect(actions.compareDocumentPosition(shortcuts) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it('keeps the pinned footer layout when primaryActions returns null without an actions element', () => {
    const Page = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      primaryActions: () => null,
      Board,
    });
    renderWithProviders(<Page />);
    expect(screen.getByTestId('game-footer-content')).toBeInTheDocument();
    expect(screen.queryByTestId('game-footer-actions')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'リセット' }).closest('[data-testid="game-footer-content"]')).toBeNull();
  });

  it('keeps the pinned footer layout when primaryActions returns undefined', () => {
    const Page = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      primaryActions: () => undefined,
      Board,
    });
    renderWithProviders(<Page />);
    expect(screen.getByTestId('game-footer-content')).toBeInTheDocument();
    expect(screen.queryByTestId('game-footer-actions')).not.toBeInTheDocument();
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

  it('omits LandscapeBanner by default and renders it in GUI mode only when enabled', async () => {
    const Page = makeDefinition();
    const defaultView = renderWithProviders(<Page />);
    expect(screen.queryByText('横画面での表示を推奨します')).not.toBeInTheDocument();
    defaultView.unmount();

    const BannerPage = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      landscapeBanner: true,
      Board,
    });
    const view = renderWithProviders(<BannerPage />);
    expect(screen.getByText('横画面での表示を推奨します')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'CLIモードに切り替え' }));
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    expect(screen.queryByText('横画面での表示を推奨します')).not.toBeInTheDocument();
    view.unmount();
  });

  it('places settings before or after the scroll content according to settingsPlacement', () => {
    const settings = () => [
      { items: [{ type: 'checkbox' as const, id: 'test-setting', label: 'Test', checked: false }] },
    ];
    const TopPage = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      settings,
      settingsPlacement: 'top',
      Board,
    });
    const { unmount } = renderWithProviders(<TopPage />);
    const topSettings = screen.getByText('設定').parentElement;
    const topBoard = screen.getByTestId('board');
    expect(topSettings).not.toBeNull();
    if (topSettings)
      expect(topSettings.compareDocumentPosition(topBoard) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    unmount();
    const BottomPage = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
      phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
      settings,
      settingsPlacement: 'bottom',
      Board,
    });
    renderWithProviders(<BottomPage />);
    const bottomSettings = screen.getByText('設定').parentElement;
    const bottomBoard = screen.getByTestId('board');
    expect(bottomSettings).not.toBeNull();
    if (bottomSettings)
      expect(bottomBoard.compareDocumentPosition(bottomSettings) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  describe('page-shaped options (#12047 batch: OldMaid)', () => {
    function makePage(over: Partial<GameDefinition<State, Game, unknown[]>>) {
      currentGame = { ...defaultBaseGame };
      return createGamePage<State, Game, unknown[]>({
        key: 'acesup',
        useGame: useFakeGame,
        tutorial: [],
        skeleton: { kind: 'centered', rows: [1] },
        cli: () => ({ parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] }),
        phase: ({ state }) => ({ name: 'playing', ended: state.ended }),
        Board,
        FooterActions,
        ...over,
      });
    }

    it('passes the frontend hint state to the cli definition', () => {
      const cli = vi.fn((_ctx: { frontendHint?: { hintEnabled: boolean } }) => ({
        parseCommand: () => ({ args: [] }),
        formatResponse: () => 'state',
        helpText: [],
      }));
      const Page = makePage({ cli });
      renderWithProviders(<Page />);
      expect(cli).toHaveBeenCalled();
      const ctx = cli.mock.calls[0]?.[0];
      expect(ctx?.frontendHint).toBeDefined();
      expect(ctx?.frontendHint?.hintEnabled).toBe(false);
    });

    it('puts footerTutorialId on the fixed footer element, only with primary actions', () => {
      const WithTutorial = makePage({
        primaryActions: () => <button type="button">Play</button>,
        footerTutorialId: 'cb-play-button',
      });
      const first = renderWithProviders(<WithTutorial />);
      expect(screen.getByRole('contentinfo')).toHaveAttribute('data-tutorial', 'cb-play-button');
      first.unmount();

      const Omitted = makePage({ primaryActions: () => <button type="button">Play</button> });
      const second = renderWithProviders(<Omitted />);
      expect(second.container.querySelector('[data-tutorial]')).toBeNull();
      second.unmount();

      const NoPrimaryActions = makePage({ footerTutorialId: 'cb-play-button' });
      const third = renderWithProviders(<NoPrimaryActions />);
      expect(third.container.querySelector('[data-tutorial="cb-play-button"]')).toBeNull();
    });

    it('renders AboveContent after top settings and outside the scrolling content', () => {
      const AboveContent = () => <div data-testid="above">above</div>;
      const Page = makePage({ AboveContent, settingsPlacement: 'top' });
      renderWithProviders(<Page />);
      const above = screen.getByTestId('above');
      const scroller = screen.getByTestId('board').closest('.overflow-y-auto');
      expect(scroller).not.toBeNull();
      expect(scroller?.contains(above)).toBe(false);
      if (scroller) expect(above.compareDocumentPosition(scroller) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
      const settings = document.querySelector('details');
      expect(settings).not.toBeNull();
      if (settings) expect(settings.compareDocumentPosition(above) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    });

    it('replaces the action row classes with actionsClassName', () => {
      const Page = makePage({ actionsClassName: 'text-center' });
      renderWithProviders(<Page />);
      const row = screen.getByRole('button', { name: 'Give up' }).parentElement;
      expect(row).toHaveClass('text-center');
      expect(row).not.toHaveClass('flex');
    });

    it("lets the page place the reset button itself when resetButton is 'custom'", () => {
      const Actions = ({ renderResetButton }: GamePageContext<State, Game>) => (
        <>
          <button type="button">First</button>
          {renderResetButton('min-w-[80px]')}
          <button type="button">Last</button>
        </>
      );
      const Page = makePage({ resetButton: 'custom', FooterActions: Actions, resetTutorialId: 'x-reset' });
      renderWithProviders(<Page />);
      const resets = screen.getAllByRole('button', { name: 'リセット' });
      expect(resets).toHaveLength(1);
      const [reset] = resets;
      expect(reset).toHaveClass('min-w-[80px]');
      expect(reset).toHaveAttribute('data-tutorial', 'x-reset');
      const buttons = Array.from(reset.parentElement?.querySelectorAll('button') ?? []).map((b) => b.textContent);
      expect(buttons).toEqual(['First', 'リセット', 'Last']);
    });

    it('lets shell override the theme background', () => {
      const Page = makePage({ shell: () => ({ gameThemeBg: 'bg-test-override' }) });
      const { container } = renderWithProviders(<Page />);
      expect(container.querySelector('.bg-test-override')).not.toBeNull();
    });
  });
});
