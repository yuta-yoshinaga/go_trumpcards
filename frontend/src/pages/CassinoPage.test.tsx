import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cassinoApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CassinoResponse } from '../types/card';
import { CassinoPage } from './CassinoPage';

vi.mock('../api/gameApi', () => ({
  cassinoApi: { exec: vi.fn() },
  actionLogApi: { cassino: vi.fn() },
}));

const mockExec = vi.mocked(cassinoApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<CassinoResponse> = {}): CassinoResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('SPADE', 3), card('HEART', 5), card('DIAMOND', 7)],
        capturedCount: 0,
        sweepCount: 0,
        totalScore: 0,
      },
      { id: 1, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
      { id: 2, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
      { id: 3, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
    ],
    currentTurn: 0,
    tableCards: [card('SPADE', 2), card('HEART', 5)],
    builds: [],
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 'playerTurn',
    config: { targetScore: 21, multiBuildEnabled: true, sweepBonusEnabled: true, cpuDifficulty: 1 },
    cpuActions: [],
    humanAction: null,
    remainingDeck: 32,
    packsDealt: 1,
    roundWinners: [],
    lastRoundDetail: null,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('CassinoPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the human hand', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-1')).toBeInTheDocument();
    expect(screen.getByTestId('hand-card-2')).toBeInTheDocument();
  });

  it('renders table cards', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('table-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('table-card-1')).toBeInTheDocument();
  });

  it('take button is disabled until both hand and table are selected', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeInTheDocument());
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('table-card-0'));
    await waitFor(() => expect(screen.getByTestId('take-button')).not.toBeDisabled());
  });

  it('calls take when Take is clicked with selections', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('take-button'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('take', {
        handIndex: 0,
        tableIndices: [0],
        buildIndices: [],
      }),
    );
  });

  it('calls trail when Trail is clicked with a hand selection', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('trail-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trail', { handIndex: 0 }));
  });

  it('calls build when Build is clicked with selection + declared value', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.change(screen.getByTestId('build-value-select'), { target: { value: '5' } });
    fireEvent.click(screen.getByTestId('build-button'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('build', {
        handIndex: 0,
        tableIndices: [0],
        declaredValue: 5,
      }),
    );
  });

  it('disables actions when it is not human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 1 }));
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeDisabled());
    expect(screen.getByTestId('trail-button')).toBeDisabled();
  });

  it('resets with config and passes it to the API', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: {
          targetScore: 21,
          multiBuildEnabled: true,
          sweepBonusEnabled: true,
          cpuDifficulty: 1,
        },
      }),
    );
  });

  it('toggles multi-build setting', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const multi = screen.getByRole('checkbox', { name: /複合ビルド|Multi-Builds/ });
    expect(multi).toBeChecked();
    fireEvent.click(multi);
    await waitFor(() => expect(multi).not.toBeChecked());
  });

  it('changes CPU difficulty', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const select = screen.getByLabelText(/CPU難易度|CPU Difficulty/);
    fireEvent.change(select, { target: { value: '2' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'reset',
        expect.objectContaining({ config: expect.objectContaining({ cpuDifficulty: 2 }) }),
      ),
    );
  });

  it('shows loading state when state has fewer than 4 players', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [{ id: 0, isHuman: true, cardCount: 0, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 }],
      }),
    );
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.queryByTestId('hand-card-0')).not.toBeInTheDocument());
  });

  it('renders builds area', async () => {
    mockExec.mockResolvedValue(
      makeState({
        builds: [{ ownerIdx: 1, value: 8, groups: [[card('SPADE', 3), card('HEART', 5)]], isMulti: false }],
      }),
    );
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('build-0')).toBeInTheDocument());
  });

  it('toggles a build selection and includes it in take', async () => {
    mockExec.mockResolvedValue(
      makeState({
        builds: [{ ownerIdx: 1, value: 8, groups: [[card('SPADE', 3), card('HEART', 5)]], isMulti: false }],
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 1,
            cards: [card('HEART', 8)],
            capturedCount: 0,
            sweepCount: 0,
            totalScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
          { id: 2, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
          { id: 3, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
        ],
        tableCards: [],
      }),
    );
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('build-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('build-0'));
    fireEvent.click(screen.getByTestId('take-button'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('take', {
        handIndex: 0,
        tableIndices: [],
        buildIndices: [0],
      }),
    );
  });

  it('toggles sweepBonusEnabled setting', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const sweep = screen.getByRole('checkbox', { name: /スイープボーナス|Sweep Bonus/ });
    expect(sweep).toBeChecked();
    fireEvent.click(sweep);
    await waitFor(() => expect(sweep).not.toBeChecked());
  });

  it('renders CLI terminal when CLI mode is enabled via localStorage', async () => {
    localStorage.setItem('cli-mode-cassino', 'true');
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    localStorage.removeItem('cli-mode-cassino');
  });
});
