import { fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { escobaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeEscobaState } from '../test/stateFactories';
import { EscobaPage } from './EscobaPage';

vi.mock('../api/gameApi', () => ({
  escobaApi: { exec: vi.fn() },
  actionLogApi: { escoba: vi.fn() },
}));

const mockExec = vi.mocked(escobaApi.exec);

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeEscobaState());
});

afterEach(async () => {
  await i18n.changeLanguage('ja');
});

describe('EscobaPage', () => {
  it('calls reset on mount with the short "r" command', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('r'));
  });

  it('renders CPU difficulty options with localized labels', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // Difficulty options are localized (ja), not the hardcoded Easy/Normal/Hard.
    expect(screen.getByRole('option', { name: 'かんたん' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'ふつう' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'むずかしい' })).toBeInTheDocument();
  });

  it('renders the human hand', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-1')).toBeInTheDocument();
    expect(screen.getByTestId('hand-card-2')).toBeInTheDocument();
  });

  it('renders table cards', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('table-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('table-card-1')).toBeInTheDocument();
  });

  it('renders per-player scores and stock', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('player-score-0')).toBeInTheDocument());
    expect(screen.getByTestId('player-score-1')).toBeInTheDocument();
    expect(screen.getByTestId('stock-remaining')).toBeInTheDocument();
  });

  it('take button is disabled until both hand and table are selected', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeInTheDocument());
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('table-card-0'));
    await waitFor(() => expect(screen.getByTestId('take-button')).not.toBeDisabled());
  });

  it('lay button is enabled when a hand card is selected and no table card', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('lay-button')).toBeInTheDocument());
    expect(screen.getByTestId('lay-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    await waitFor(() => expect(screen.getByTestId('lay-button')).not.toBeDisabled());
  });

  it('plays "p" with sorted table indices when Take is clicked', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-1'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('take-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: 0, tableIndices: [0, 1] }));
  });

  it('plays "p" with empty table indices when Lay is clicked', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('lay-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: 0, tableIndices: [] }));
  });

  it('disables actions when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeEscobaState({ currentTurn: 1, isHumanTurn: false }));
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeDisabled());
    expect(screen.getByTestId('lay-button')).toBeDisabled();
  });

  it('shows the round-end breakdown and a next-round button on roundEnd', async () => {
    mockExec.mockResolvedValue(
      makeEscobaState({
        phase: 'roundEnd',
        isHumanTurn: false,
        lastRoundDetail: {
          cards: [1, 0, 0, 0],
          espadas: [1, 0, 0, 0],
          sevens: [0, 1, 0, 0],
          oros: [1, 0, 0, 0],
          escobas: [1, 0, 0, 0],
          gained: [4, 1, 0, 0],
          aceEspada: 0,
          seteEspada: 1,
        },
      }),
    );
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('round-detail')).toBeInTheDocument());
    expect(screen.getByTestId('next-round-button')).toBeInTheDocument();
  });

  it('next-round button dispatches "n"', async () => {
    mockExec.mockResolvedValue(makeEscobaState({ phase: 'roundEnd', isHumanTurn: false }));
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('next-round-button')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('next-round-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('n'));
  });

  it('resets with config and passes it to the API', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('r', {
        config: { targetScore: 10, cpuDifficulty: 1 },
      }),
    );
  });

  it('changes CPU difficulty and includes it in the reset config', async () => {
    renderWithProviders(<EscobaPage />);
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

  it('shows loading state when state has fewer than 4 players', async () => {
    mockExec.mockResolvedValue(
      makeEscobaState({
        players: [{ id: 0, isHuman: true, handCount: 0, cards: [], capturedCount: 0, escobaCount: 0, score: 0 }],
      }),
    );
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.queryByTestId('hand-card-0')).not.toBeInTheDocument());
  });

  it('renders CLI terminal when CLI mode is enabled via localStorage', async () => {
    localStorage.setItem('cli-mode-escoba', 'true');
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    localStorage.removeItem('cli-mode-escoba');
  });
});
