import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { doudizhuApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { DoudizhuResponse } from '../types/card';
import { DoudizhuPage, formatDDZState } from './DoudizhuPage';

vi.mock('../api/gameApi', () => ({
  doudizhuApi: { exec: vi.fn() },
  actionLogApi: { doudizhu: vi.fn() },
}));

const mockExec = vi.mocked(doudizhuApi.exec);

const defaultState: DoudizhuResponse = {
  players: [
    { id: 0, isHuman: true, isFinished: false, isLandlord: true, cardCount: 20, cards: [] },
    { id: 1, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
    { id: 2, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
  ],
  phase: 'play',
  currentTurn: 0,
  tableCards: [],
  tableCombo: '',
  kittyCards: [],
  landlordIdx: 0,
  baseBid: 1,
  highestBid: 1,
  bombCount: 0,
  scores: [0, 0, 0],
  gameEndFlag: false,
  config: { cpuDifficulty: 0 },
  cpuActions: [],
  humanAction: null,
  message: '',
};

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(defaultState);
});

describe('DoudizhuPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<DoudizhuPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset command on mount', async () => {
    renderWithProviders(<DoudizhuPage />);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'reset' }));
    });
  });

  it('renders CPU player areas after load', async () => {
    renderWithProviders(<DoudizhuPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1/)).toBeInTheDocument();
    });
  });

  it('renders result message when game ends', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      phase: 'end',
      gameEndFlag: true,
      scores: [2, -1, -1],
    });
    renderWithProviders(<DoudizhuPage />);
    await waitFor(() => {
      expect(screen.getByText('地主の勝利！')).toBeInTheDocument();
    });
  });

  it('renders peasant-win message when landlord loses', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      phase: 'end',
      gameEndFlag: true,
      scores: [-2, 1, 1],
    });
    renderWithProviders(<DoudizhuPage />);
    await waitFor(() => {
      expect(screen.getByText('農民の勝利！')).toBeInTheDocument();
    });
  });

  it('renders bid buttons above the highest bid during bid phase', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      phase: 'bid',
      landlordIdx: -1,
      highestBid: 1,
      currentTurn: 0,
    });
    renderWithProviders(<DoudizhuPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '2で叫ぶ' })).toBeInTheDocument();
    });
    // Bid 1 is not offered because highestBid is already 1.
    expect(screen.queryByRole('button', { name: '1で叫ぶ' })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '2で叫ぶ' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'bid', bidValue: 2 }));
    });
  });

  it('shows the human hand during the bid phase as display-only (no selection)', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      phase: 'bid',
      landlordIdx: -1,
      highestBid: 0,
      currentTurn: 0,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          isLandlord: false,
          cardCount: 1,
          cards: [{ design: 'SPADE', value: 5 }],
        },
        { id: 1, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
        { id: 2, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
      ],
    });
    renderWithProviders(<DoudizhuPage />);

    // The hand card is rendered during the bid phase.
    const cardBtn = await screen.findByRole('button', { name: '♠ 5' });
    // It is display-only: disabled, with no selection toggle state.
    expect(cardBtn).toBeDisabled();
    expect(cardBtn).not.toHaveAttribute('aria-pressed');
    fireEvent.click(cardBtn);
    expect(cardBtn).not.toHaveAttribute('aria-pressed');

    // No play/pass controls during bidding.
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('passes during the bid phase', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      phase: 'bid',
      landlordIdx: -1,
      highestBid: 0,
      currentTurn: 0,
    });
    renderWithProviders(<DoudizhuPage />);
    const passButton = await screen.findByRole('button', { name: 'パス' });
    fireEvent.click(passButton);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'bid', bidValue: 0 }));
    });
  });

  it('formatDDZState lists the human hand with indices for the CLI', () => {
    const out = formatDDZState({
      ...defaultState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          isLandlord: true,
          cardCount: 2,
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'HEART', value: 13 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
        { id: 2, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
      ],
    });
    expect(out).toContain('[0]♠ 5');
    expect(out).toContain('[1]♥ K');
  });

  it('formatDDZState omits the hand line when the human has no visible cards', () => {
    // defaultState's human has cards: [] → no "Your hand" line.
    expect(formatDDZState(defaultState)).not.toContain('Your hand');
  });

  it('labels hand cards via cardAlt and reflects selection with aria-pressed', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          isLandlord: true,
          cardCount: 1,
          cards: [{ design: 'SPADE', value: 5 }],
        },
        { id: 1, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
        { id: 2, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
      ],
    });
    renderWithProviders(<DoudizhuPage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ 5' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
  });

  it('disables play until a card is selected, then plays on click', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          isLandlord: true,
          cardCount: 1,
          cards: [{ design: 'SPADE', value: 5 }],
        },
        { id: 1, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
        { id: 2, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
      ],
    });
    renderWithProviders(<DoudizhuPage />);

    const playButton = await screen.findByRole('button', { name: '出す' });
    expect(playButton).toBeDisabled();

    fireEvent.click(screen.getByAltText(/5/));
    expect(playButton).toBeEnabled();

    fireEvent.click(playButton);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'p', indices: [0] }));
    });
  });

  it('toggles card selection off and passes when the table has cards', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      tableCards: [{ design: 'CLOVER', value: 13 }],
      tableCombo: 'single',
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          isLandlord: false,
          cardCount: 1,
          cards: [{ design: 'SPADE', value: 5 }],
        },
        { id: 1, isHuman: false, isFinished: false, isLandlord: true, cardCount: 17, cards: [] },
        { id: 2, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
      ],
    });
    renderWithProviders(<DoudizhuPage />);

    const playButton = await screen.findByRole('button', { name: '出す' });
    // Select then deselect — play returns to disabled (toggleCard delete branch).
    fireEvent.click(screen.getByAltText('♠ 5'));
    expect(playButton).toBeEnabled();
    fireEvent.click(screen.getByAltText('♠ 5'));
    expect(playButton).toBeDisabled();

    // Pass is available because the table is non-empty.
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'p', indices: [] }));
    });
  });

  it('warns that a mismatched selection is not a valid combo', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          isLandlord: true,
          cardCount: 2,
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'HEART', value: 9 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
        { id: 2, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
      ],
    });
    renderWithProviders(<DoudizhuPage />);
    fireEvent.click(await screen.findByRole('button', { name: '♠ 5' }));
    fireEvent.click(screen.getByRole('button', { name: '♥ 9' }));
    expect(await screen.findByTestId('ddz-invalid-combo')).toBeInTheDocument();
  });

  it('warns that a valid-but-too-low combo cannot beat the table', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      tableCards: [
        { design: 'SPADE', value: 8 },
        { design: 'HEART', value: 8 },
        { design: 'DIAMOND', value: 8 },
      ],
      tableCombo: 'trio',
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          isLandlord: false,
          cardCount: 3,
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'HEART', value: 5 },
            { design: 'DIAMOND', value: 5 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, isLandlord: true, cardCount: 17, cards: [] },
        { id: 2, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
      ],
    });
    renderWithProviders(<DoudizhuPage />);
    fireEvent.click(await screen.findByRole('button', { name: '♠ 5' }));
    fireEvent.click(screen.getByRole('button', { name: '♥ 5' }));
    fireEvent.click(screen.getByRole('button', { name: '♦ 5' }));
    expect(await screen.findByTestId('ddz-no-beat')).toBeInTheDocument();
  });

  it('shows the combo type for a valid beating play', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      tableCards: [{ design: 'CLOVER', value: 13 }],
      tableCombo: 'single',
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          isLandlord: false,
          cardCount: 1,
          cards: [{ design: 'SPADE', value: 2 }],
        },
        { id: 1, isHuman: false, isFinished: false, isLandlord: true, cardCount: 17, cards: [] },
        { id: 2, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
      ],
    });
    renderWithProviders(<DoudizhuPage />);
    fireEvent.click(await screen.findByRole('button', { name: '♠ 2' }));
    const badge = await screen.findByTestId('ddz-combo-type');
    // "選択中: 単張（1枚）" — single, and no warning is shown.
    expect(badge).toHaveTextContent('単張');
    expect(screen.queryByTestId('ddz-invalid-combo')).not.toBeInTheDocument();
    expect(screen.queryByTestId('ddz-no-beat')).not.toBeInTheDocument();
  });

  it('opens the reset confirmation and resets on confirm', async () => {
    renderWithProviders(<DoudizhuPage />);
    const resetButton = await screen.findByRole('button', { name: 'リセット' });
    fireEvent.click(resetButton);
    const confirmButton = await screen.findByRole('button', { name: '確認' });
    fireEvent.click(confirmButton);
    await waitFor(() => {
      // reset is also called on mount; confirm triggers another reset.
      expect(mockExec.mock.calls.filter((c) => c[0]?.command === 'reset').length).toBeGreaterThan(1);
    });
  });

  it('renders kitty cards and the current table combo', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      kittyCards: [
        { design: 'SPADE', value: 3 },
        { design: 'HEART', value: 4 },
        { design: 'DIAMOND', value: 5 },
      ],
      tableCards: [{ design: 'CLOVER', value: 13 }],
      tableCombo: 'single',
    });
    renderWithProviders(<DoudizhuPage />);

    await waitFor(() => {
      expect(screen.getByText('single')).toBeInTheDocument();
    });
    // Kitty label and the three kitty card images are rendered.
    expect(screen.getByText(/底牌/)).toBeInTheDocument();
    expect(screen.getByAltText('♠ 3')).toBeInTheDocument();
  });

  it('renders the CLI terminal when CLI mode is enabled', async () => {
    localStorage.setItem('cli-mode-doudizhu', 'true');
    renderWithProviders(<DoudizhuPage />);
    await waitFor(() => {
      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });
  });

  it('parses and dispatches CLI commands, and formats state output', async () => {
    localStorage.setItem('cli-mode-doudizhu', 'true');
    mockExec.mockResolvedValue({
      ...defaultState,
      tableCards: [{ design: 'CLOVER', value: 13 }],
      tableCombo: 'single',
      message: 'your turn',
    });
    renderWithProviders(<DoudizhuPage />);
    const input = await screen.findByRole('textbox');

    const run = async (cmd: string, expected: Record<string, unknown>) => {
      fireEvent.change(input, { target: { value: cmd } });
      fireEvent.keyDown(input, { key: 'Enter' });
      await waitFor(() => {
        expect(mockExec).toHaveBeenCalledWith(expect.objectContaining(expected));
      });
    };

    await run('p 0 1', { command: 'p', indices: [0, 1] });
    await run('bid 2', { command: 'bid', bidValue: 2 });
    await run('r', { command: 'reset' });
    await run('sd 1', { command: 'reset', config: { cpuDifficulty: 1 } });
    await run('log', { command: 'l' });

    // Unknown command surfaces an error and does not dispatch.
    mockExec.mockClear();
    fireEvent.change(input, { target: { value: 'xyz' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    await waitFor(() => {
      expect(screen.getAllByText(/Unknown command/).length).toBeGreaterThan(0);
    });
    expect(mockExec).not.toHaveBeenCalled();
  });
});
