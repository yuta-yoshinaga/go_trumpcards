import { fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { scoponeApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeScoponeState } from '../test/stateFactories';
import { ScoponePage } from './ScoponePage';

vi.mock('../api/gameApi', () => ({
  scoponeApi: { exec: vi.fn() },
  actionLogApi: { scopone: vi.fn() },
}));

const mockExec = vi.mocked(scoponeApi.exec);

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeScoponeState());
});

afterEach(async () => {
  await i18n.changeLanguage('ja');
});

describe('ScoponePage', () => {
  it('renders the GameSkeleton while the initial state is loading', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ScoponePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the short "r" command', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('r'));
  });

  it('renders CPU difficulty options with localized labels', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // Difficulty options are localized (ja), not the hardcoded Easy/Normal/Hard.
    expect(screen.getByRole('option', { name: 'かんたん' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'ふつう' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'むずかしい' })).toBeInTheDocument();
  });

  it('renders the human hand', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-1')).toBeInTheDocument();
    expect(screen.getByTestId('hand-card-2')).toBeInTheDocument();
  });

  it('renders table cards', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('table-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('table-card-1')).toBeInTheDocument();
  });

  it('renders team scores', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('team-score-0')).toBeInTheDocument());
    expect(screen.getByTestId('team-score-1')).toBeInTheDocument();
  });

  it('take button is disabled until both hand and table are selected', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeInTheDocument());
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('table-card-0'));
    await waitFor(() => expect(screen.getByTestId('take-button')).not.toBeDisabled());
  });

  it('lay button is enabled when a hand card is selected and no table card', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('lay-button')).toBeInTheDocument());
    expect(screen.getByTestId('lay-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    await waitFor(() => expect(screen.getByTestId('lay-button')).not.toBeDisabled());
  });

  it('plays "p" with sorted table indices when Take is clicked', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-1'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('take-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: 0, tableIndices: [0, 1] }));
  });

  it('plays "p" with empty table indices when Lay is clicked', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('lay-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: 0, tableIndices: [] }));
  });

  it('disables actions when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeScoponeState({ currentTurn: 1, isHumanTurn: false }));
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeDisabled());
    expect(screen.getByTestId('lay-button')).toBeDisabled();
  });

  it('shows the round-end breakdown and a next-round button on roundEnd', async () => {
    mockExec.mockResolvedValue(
      makeScoponeState({
        phase: 'roundEnd',
        isHumanTurn: false,
        lastRoundDetail: {
          cards: [1, 0],
          diamonds: [1, 0],
          sevens: [0, 1],
          scopas: [1, 0],
          gained: [3, 1],
          settebello: 0,
        },
      }),
    );
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('round-detail')).toBeInTheDocument());
    expect(screen.getByTestId('next-round-button')).toBeInTheDocument();
  });

  it('next-round button dispatches "n"', async () => {
    mockExec.mockResolvedValue(makeScoponeState({ phase: 'roundEnd', isHumanTurn: false }));
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('next-round-button')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('next-round-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('n'));
  });

  it('resets with config and passes it to the API', async () => {
    renderWithProviders(<ScoponePage />);
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
    renderWithProviders(<ScoponePage />);
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
      makeScoponeState({
        players: [{ id: 0, isHuman: true, team: 0, handCount: 0, cards: [], capturedCount: 0, scopaCount: 0 }],
      }),
    );
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.queryByTestId('hand-card-0')).not.toBeInTheDocument());
  });

  it('renders CLI terminal when CLI mode is enabled via localStorage', async () => {
    localStorage.setItem('cli-mode-scopone', 'true');
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    localStorage.removeItem('cli-mode-scopone');
  });
});
