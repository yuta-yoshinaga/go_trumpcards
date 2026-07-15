import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { openfacechineseApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, OpenFaceChinesePlayer, OpenFaceChineseResponse } from '../types/card';
import { OpenFaceChinesePage } from './OpenFaceChinesePage';

vi.mock('../api/gameApi', () => ({
  openfacechineseApi: { exec: vi.fn() },
  actionLogApi: { openfacechinese: vi.fn() },
}));

const cliModeDisabled = {
  cliEnabled: false,
  toggleCli: vi.fn(),
  logEntries: [],
  addInput: vi.fn(),
  addOutput: vi.fn(),
  addError: vi.fn(),
  clearLog: vi.fn(),
};
vi.mock('../hooks/useCliMode', () => ({ useCliMode: vi.fn() }));

const mockExec = vi.mocked(openfacechineseApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makePlayer(overrides: Partial<OpenFaceChinesePlayer> = {}): OpenFaceChinesePlayer {
  return {
    id: 0,
    isHuman: true,
    front: [],
    middle: [],
    back: [],
    pending: [],
    roundScore: 0,
    royalty: 0,
    fouled: false,
    fantasyland: false,
    totalScore: 0,
    ...overrides,
  };
}

function makeState(overrides: Partial<OpenFaceChineseResponse> = {}): OpenFaceChineseResponse {
  return {
    phase: 0,
    roundNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    currentCard: card('SPADE', 1),
    gameEndFlag: false,
    winnerIdx: -1,
    isHumanTurn: true,
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false })],
    config: { cpuDifficulty: 0, playerCount: 2, targetRounds: 1 },
    message: '',
    ...overrides,
  };
}

const placingState = makeState();
const roundEndState = makeState({
  phase: 1,
  isHumanTurn: false,
  currentCard: undefined,
  players: [
    makePlayer({
      front: [card('SPADE', 2), card('HEART', 3), card('CLOVER', 4)],
      middle: [card('SPADE', 5), card('HEART', 6), card('CLOVER', 7), card('DIAMOND', 8), card('SPADE', 9)],
      back: [card('SPADE', 10), card('HEART', 11), card('CLOVER', 12), card('DIAMOND', 13), card('SPADE', 1)],
      roundScore: 6,
      royalty: 4,
      totalScore: 6,
    }),
    makePlayer({ id: 1, isHuman: false, roundScore: -6, fouled: true, totalScore: -6 }),
  ],
});
const gameEndState = makeState({
  phase: 2,
  gameEndFlag: true,
  isHumanTurn: false,
  winnerIdx: 0,
  currentCard: undefined,
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(placingState);
  vi.mocked(useCliMode).mockReturnValue(cliModeDisabled);
});

describe('OpenFaceChinesePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<OpenFaceChinesePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<OpenFaceChinesePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the round number and the player boards', async () => {
    renderWithProviders(<OpenFaceChinesePage />);
    await waitFor(() => expect(screen.getByTestId('player-0')).toBeInTheDocument());
    expect(screen.getByTestId('player-1')).toBeInTheDocument();
    expect(screen.getByText(/ラウンド: 1/)).toBeInTheDocument();
  });

  it('renders the three place buttons during placing', async () => {
    renderWithProviders(<OpenFaceChinesePage />);
    await waitFor(() => expect(screen.getByTestId('place-front')).toBeInTheDocument());
    expect(screen.getByTestId('place-middle')).toBeInTheDocument();
    expect(screen.getByTestId('place-back')).toBeInTheDocument();
  });

  it('announces the pending card and labels rows for screen readers', async () => {
    renderWithProviders(<OpenFaceChinesePage />); // currentCard ♠A, human placing
    const announce = await screen.findByTestId('ofc-pending-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    expect(announce).toHaveTextContent('配置待ち: ♠ A');
    // Each row is a named group; both players' empty top rows report their count.
    expect(screen.getAllByRole('group', { name: /トップ（3枚） 0枚/ }).length).toBeGreaterThanOrEqual(1);
  });

  it('names a filled row with its card contents', async () => {
    mockExec.mockResolvedValue(roundEndState); // human top row = ♠2 ♥3 ♣4
    renderWithProviders(<OpenFaceChinesePage />);
    await waitFor(() =>
      expect(screen.getByRole('group', { name: 'トップ（3枚） 3枚: ♠ 2, ♥ 3, ♣ 4' })).toBeInTheDocument(),
    );
  });

  it('places into the top row', async () => {
    renderWithProviders(<OpenFaceChinesePage />);
    const btn = await screen.findByTestId('place-front');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('place', { row: 0 }));
  });

  it('places into the middle row', async () => {
    renderWithProviders(<OpenFaceChinesePage />);
    const btn = await screen.findByTestId('place-middle');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('place', { row: 1 }));
  });

  it('places into the bottom row', async () => {
    renderWithProviders(<OpenFaceChinesePage />);
    const btn = await screen.findByTestId('place-back');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('place', { row: 2 }));
  });

  it('hides place buttons when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ isHumanTurn: false, currentCard: undefined }));
    renderWithProviders(<OpenFaceChinesePage />);
    await waitFor(() => expect(screen.getByTestId('player-0')).toBeInTheDocument());
    expect(screen.queryByTestId('place-front')).not.toBeInTheDocument();
  });

  it('renders round-end scores, royalty and foul, then dispatches nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<OpenFaceChinesePage />);
    await waitFor(() => expect(screen.getByTestId('round-score-0')).toBeInTheDocument());
    expect(screen.getByText(/ロイヤリティ: \+4/)).toBeInTheDocument();
    expect(screen.getByText('ファウル（無効な並び）')).toBeInTheDocument();
    const nextBtn = screen.getByTestId('next-button');
    mockExec.mockClear();
    fireEvent.click(nextBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows the game-over label at game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<OpenFaceChinesePage />);
    await waitFor(() => expect(screen.getAllByText('ゲーム終了').length).toBeGreaterThan(0));
    expect(screen.queryByTestId('next-button')).not.toBeInTheDocument();
  });

  it('renders the CLI terminal and hides the board when CLI mode is enabled', async () => {
    vi.mocked(useCliMode).mockReturnValue({ ...cliModeDisabled, cliEnabled: true });
    mockExec.mockResolvedValue(placingState);
    renderWithProviders(<OpenFaceChinesePage />);
    await waitFor(() => expect(screen.getByRole('log')).toBeInTheDocument());
    expect(screen.queryByTestId('place-front')).not.toBeInTheDocument();
  });
});
