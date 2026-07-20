import { fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { scopaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, ScopaResponse } from '../types/card';
import { ScopaPage } from './ScopaPage';

vi.mock('../api/gameApi', () => ({
  scopaApi: { exec: vi.fn() },
  actionLogApi: { scopa: vi.fn() },
}));

const mockExec = vi.mocked(scopaApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<ScopaResponse> = {}): ScopaResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('SPADE', 3), card('HEART', 5), card('DIAMOND', 7)],
        capturedCount: 0,
        scopaCount: 0,
        totalScore: 0,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], capturedCount: 0, scopaCount: 0, totalScore: 0 },
    ],
    currentTurn: 0,
    tableCards: [card('SPADE', 2), card('HEART', 5)],
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 'playerTurn',
    config: { targetScore: 11, cpuDifficulty: 1 },
    cpuActions: [],
    humanAction: null,
    remainingDeck: 30,
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

describe('ScopaPage', () => {
  it('calls reset on mount with the short "r" command', async () => {
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('r'));
  });

  it('renders the human hand', async () => {
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-1')).toBeInTheDocument();
    expect(screen.getByTestId('hand-card-2')).toBeInTheDocument();
  });

  it('renders table cards', async () => {
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('table-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('table-card-1')).toBeInTheDocument();
  });

  it('exposes capture candidates, selection state, and a candidate-count live region', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('table-card-0')).toBeInTheDocument());
    // Base state: no hand card selected → plain labels, not pressed, empty live region.
    expect(screen.getByTestId('table-card-0')).toHaveAttribute('aria-label', '♠ 2');
    expect(screen.getByTestId('table-card-0')).toHaveAttribute('aria-pressed', 'false');
    expect(screen.getByTestId('sc-take-candidate-live')).toHaveTextContent('');
    // Selecting the ♥5 hand card makes the matching ♥5 table card a capture candidate.
    fireEvent.click(screen.getByTestId('hand-card-1'));
    await waitFor(() => expect(screen.getByTestId('table-card-1')).toHaveAttribute('aria-label', '♥ 5 取り札候補'));
    expect(screen.getByTestId('sc-take-candidate-live')).toHaveTextContent('取り札候補があります');
    // Selecting that candidate flips its aria-pressed to true and updates the label.
    fireEvent.click(screen.getByTestId('table-card-1'));
    await waitFor(() => expect(screen.getByTestId('table-card-1')).toHaveAttribute('aria-pressed', 'true'));
    expect(screen.getByTestId('table-card-1')).toHaveAttribute('aria-label', '♥ 5 選択中');
  });

  it('renders human stats via i18n (no hardcoded Japanese under en locale)', async () => {
    await i18n.changeLanguage('en');
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    // English locale must show English stat labels, never the Japanese "枚" unit.
    expect(screen.getByText(/Hand 3 \/ Captured 0 \/ Scopa 0 \/ Total 0 pt/)).toBeInTheDocument();
    expect(screen.queryByText(/枚/)).not.toBeInTheDocument();
  });

  afterEach(async () => {
    await i18n.changeLanguage('ja');
  });

  it('take button is disabled until both hand and table are selected', async () => {
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeInTheDocument());
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('table-card-0'));
    await waitFor(() => expect(screen.getByTestId('take-button')).not.toBeDisabled());
  });

  it('lay button is enabled when a hand card is selected and no table card', async () => {
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('lay-button')).toBeInTheDocument());
    expect(screen.getByTestId('lay-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    await waitFor(() => expect(screen.getByTestId('lay-button')).not.toBeDisabled());
  });

  it('plays "p" with sorted table indices when Take is clicked', async () => {
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-1'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('take-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: 0, tableIndices: [0, 1] }));
  });

  it('plays "p" with empty table indices when Lay is clicked', async () => {
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('lay-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: 0, tableIndices: [] }));
  });

  it('disables actions when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 1 }));
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeDisabled());
    expect(screen.getByTestId('lay-button')).toBeDisabled();
  });

  it('resets with config and passes it to the API', async () => {
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('r', {
        config: { targetScore: 11, cpuDifficulty: 1 },
      }),
    );
  });

  it('changes CPU difficulty and includes it in the reset config', async () => {
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const select = screen.getByLabelText(/CPU難易度|CPU Difficulty/);
    fireEvent.change(select, { target: { value: '2' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'r',
        expect.objectContaining({ config: expect.objectContaining({ cpuDifficulty: 2 }) }),
      ),
    );
  });

  it('changes target score and includes it in the reset config', async () => {
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const select = screen.getByLabelText(/目標点|Target Score/);
    fireEvent.change(select, { target: { value: '21' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'r',
        expect.objectContaining({ config: expect.objectContaining({ targetScore: 21 }) }),
      ),
    );
  });

  it('hides the next-round button and round-end banner during normal play', async () => {
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.queryByTestId('next-round-button')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sc-round-end-banner')).not.toBeInTheDocument();
  });

  it('shows the round-end banner and a working next-round button when the round ends', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 'roundEnd', currentTurn: 1 }));
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('sc-round-end-banner')).toBeInTheDocument());
    const nextRound = screen.getByTestId('next-round-button');
    expect(nextRound).toBeInTheDocument();
    expect(nextRound).not.toBeDisabled();

    mockExec.mockClear();
    fireEvent.click(nextRound);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('n'));
  });

  it('hides the score breakdown during normal play', async () => {
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.queryByTestId('sc-score-breakdown')).not.toBeInTheDocument();
  });

  it('renders the per-category score breakdown at round end', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'roundEnd',
        currentTurn: 1,
        lastRoundDetail: {
          cards: { 0: 20, 1: 16 },
          diamonds: { 0: 4, 1: 6 },
          sevens: { 0: 3, 1: 1 },
          hasSetteBello: 0,
          scopas: { 0: 2, 1: 0 },
          gained: { 0: 5, 1: 1 },
        },
      }),
    );
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('sc-score-breakdown')).toBeInTheDocument());
    // Carte (most cards) and settebello go to the human; denari goes to CPU 1.
    expect(screen.getByTestId('sc-breakdown-cards')).toHaveTextContent('あなた +1');
    expect(screen.getByTestId('sc-breakdown-denari')).toHaveTextContent('CPU 1 +1');
    expect(screen.getByTestId('sc-breakdown-primiera')).toHaveTextContent('あなた +1');
    expect(screen.getByTestId('sc-breakdown-settebello')).toHaveTextContent('あなた +1');
    // Scopa sweeps and per-player round totals are listed too.
    expect(screen.getByTestId('sc-breakdown-scopa')).toHaveTextContent('あなた ×2');
    expect(screen.getByTestId('sc-breakdown-gained')).toHaveTextContent('あなた +5');
    expect(screen.getByTestId('sc-breakdown-gained')).toHaveTextContent('CPU 1 +1');
  });

  it('marks a tied category as no-points in the breakdown', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'roundEnd',
        currentTurn: 1,
        lastRoundDetail: {
          cards: { 0: 18, 1: 18 },
          diamonds: { 0: 0, 1: 0 },
          sevens: { 0: 0, 1: 0 },
          hasSetteBello: -1,
          scopas: { 0: 0, 1: 0 },
          gained: { 0: 0, 1: 0 },
        },
      }),
    );
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByTestId('sc-breakdown-cards')).toBeInTheDocument());
    expect(screen.getByTestId('sc-breakdown-cards')).toHaveTextContent('引き分け（点なし）');
    expect(screen.getByTestId('sc-breakdown-settebello')).toHaveTextContent('引き分け（点なし）');
  });

  it('shows loading state when state has fewer than 2 players', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [{ id: 0, isHuman: true, cardCount: 0, cards: [], capturedCount: 0, scopaCount: 0, totalScore: 0 }],
      }),
    );
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.queryByTestId('hand-card-0')).not.toBeInTheDocument());
  });

  it('renders CLI terminal when CLI mode is enabled via localStorage', async () => {
    localStorage.setItem('cli-mode-scopa', 'true');
    renderWithProviders(<ScopaPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    localStorage.removeItem('cli-mode-scopa');
  });
});
