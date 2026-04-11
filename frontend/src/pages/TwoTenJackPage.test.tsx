import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, twoTenJackApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTwoTenJackState } from '../test/stateFactories';
import { TwoTenJackPage } from './TwoTenJackPage';

vi.mock('../api/gameApi', () => ({
  twoTenJackApi: { exec: vi.fn() },
  actionLogApi: { twotenjack: vi.fn() },
}));

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockExec = vi.mocked(twoTenJackApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const playPhaseState = makeTwoTenJackState();

const declarePhaseState = makeTwoTenJackState({
  phase: 0,
  declarerIdx: 0,
  trumpSuit: -1,
});

const declarePhaseCpuState = makeTwoTenJackState({
  phase: 0,
  declarerIdx: 1,
  trumpSuit: -1,
});

const trickEndState = makeTwoTenJackState({
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
    { playerIdx: 2, card: { design: 'CLOVER', value: 7 } },
    { playerIdx: 3, card: { design: 'SPADE', value: 9 } },
  ],
});

const roundEndState = makeTwoTenJackState({ phase: 3 });

const gameEndState = makeTwoTenJackState({
  phase: 4,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'Game end!',
});

const cpuTurnState = makeTwoTenJackState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  });
});

describe('TwoTenJackPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TwoTenJackPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 50,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2660 A')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 J')).toBeInTheDocument();
    });
  });

  it('shows four suit buttons during human declare phase', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '\u30b9\u30da\u30fc\u30c9' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u30af\u30e9\u30d6' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u30cf\u30fc\u30c8' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u30c0\u30a4\u30e4' })).toBeInTheDocument();
    });
  });

  it('dispatches declare command when a suit button is clicked', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30cf\u30fc\u30c8' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30cf\u30fc\u30c8' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 3));
  });

  it('does not show suit buttons during CPU declare turn', async () => {
    mockExec.mockResolvedValue(declarePhaseCpuState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getAllByText(/CPU 1/).length).toBeGreaterThan(0));
    expect(screen.queryByRole('button', { name: '\u30cf\u30fc\u30c8' })).not.toBeInTheDocument();
  });

  it('play button disabled when no card selected', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u51fa\u3059' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '\u51fa\u3059' })).not.toBeDisabled();
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getAllByText(/CPU 1/).length).toBeGreaterThan(0));
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  it('shows next trick button on trick end', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30c8\u30ea\u30c3\u30af' })).toBeInTheDocument(),
    );
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' })).toBeInTheDocument(),
    );
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
    });
  });

  it('reset confirm dispatches reset with current config', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 50,
      }),
    );
  });

  it('shows error alert on failed reset', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('ensures actionLogApi.twotenjack is registered', () => {
    expect(actionLogApi.twotenjack).toBeDefined();
  });

  it('renders CLI terminal when CLI mode is enabled', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  it('changing cpuDifficulty updates config passed to reset', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 2,
        pointLimit: 50,
      }),
    );
  });

  it('changing pointLimit updates config passed to reset', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '100' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
      }),
    );
  });
});
