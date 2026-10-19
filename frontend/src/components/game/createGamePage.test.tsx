import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi } from '../../api/gameApi';
import { renderWithProviders } from '../../test/renderWithProviders';
import { createGamePage, type GameDefinition, type GameHookBase, type GamePageContext } from './createGamePage';

vi.mock('../../api/gameApi', () => ({ actionLogApi: { acesup: vi.fn() } }));

interface State {
  ended: boolean;
}
interface Game extends GameHookBase<State> {
  exec: (...args: never[]) => Promise<unknown>;
}
const reset = vi.fn();
const giveUp = vi.fn();
const shortcut = vi.fn();
const retry = vi.fn();

let currentGame: Game;
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
    state: { ended: false },
    loading: false,
    error: null,
    hintError: null,
    retry,
    handleReset: reset,
    exec: async () => undefined,
    ...(options.includeGiveUp ? { handleGiveUp: giveUp } : {}),
  };
  currentGame = base;
  const definition: GameDefinition<State, Game, never[]> = {
    key: 'acesup',
    useGame: useFakeGame,
    tutorial: [],
    skeleton: { kind: 'centered', rows: [1] },
    cli: { parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] },
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
  localStorage.removeItem('cli-mode-acesup');
});

describe('createGamePage', () => {
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

  it('fires shortcut keys only when enabled and not busy', () => {
    const Page = makeDefinition({ shortcuts: true });
    const view = renderWithProviders(<Page />);
    fireEvent.keyDown(document, { key: 'x' });
    expect(shortcut).toHaveBeenCalledTimes(1);
    currentGame = { ...currentGame, loading: true };
    view.rerender(<Page />);
    fireEvent.keyDown(document, { key: 'x' });
    expect(shortcut).toHaveBeenCalledTimes(1);
  });

  it('suppresses shortcuts when the definition disables them', () => {
    const DisabledPage = createGamePage({
      key: 'acesup',
      useGame: useFakeGame,
      tutorial: [],
      skeleton: { kind: 'centered', rows: [1] },
      cli: { parseCommand: () => ({ args: [] }), formatResponse: () => 'state', helpText: ['help'] },
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

  it('makes give-up a no-op when the hook has no give-up action', () => {
    const Page = makeDefinition();
    renderWithProviders(<Page />);
    fireEvent.click(screen.getByRole('button', { name: 'Give up' }));
    expect(giveUp).not.toHaveBeenCalled();
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

  it('shows hintError when the game error is null', () => {
    currentGame = { ...currentGame, hintError: 'Hint unavailable' };
    const Page = makeDefinition();
    currentGame = { ...currentGame, hintError: 'Hint unavailable' };
    renderWithProviders(<Page />);
    expect(screen.getByRole('alert')).toHaveTextContent('Hint unavailable');
  });
});
