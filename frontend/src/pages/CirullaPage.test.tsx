import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cirullaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeCirullaState } from '../test/stateFactories';
import { CirullaPage } from './CirullaPage';

vi.mock('../api/gameApi', () => ({
  cirullaApi: { exec: vi.fn() },
  actionLogApi: { cirulla: vi.fn() },
}));

const mockExec = vi.mocked(cirullaApi.exec);

const playState = makeCirullaState();

/** Selects the human's card at the given hand index. */
async function pickHand(idx: number) {
  const cards = await screen.findAllByRole('button', { name: /♠|♥|♦|♣/ });
  fireEvent.click(cards[idx]);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('CirullaPage', () => {
  it('calls reset on mount with the configured target', async () => {
    renderWithProviders(<CirullaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetScore: 51 } }),
    );
  });

  it('shows the round and the stock', async () => {
    renderWithProviders(<CirullaPage />);
    expect(await screen.findByText('ラウンド 1（51 点勝負）')).toBeInTheDocument();
    expect(screen.getByTestId('cirulla-deck')).toHaveTextContent('30');
  });

  // **場札には番号が要る。** 取る札はこの番号で指す。
  it('numbers the table cards', async () => {
    renderWithProviders(<CirullaPage />);
    const table = await screen.findByTestId('cirulla-table');
    for (const n of ['0', '1', '2', '3']) {
      expect(table).toHaveTextContent(n);
    }
  });

  // **取れる組はサーバが数えたものだけ。** 3 つの規則が絡むので、画面側で
  // 組み直すと必ずずれる。
  it('offers exactly the capture groups the server sent', async () => {
    renderWithProviders(<CirullaPage />);
    await pickHand(0); // ♠3 → [[2]]
    const box = await screen.findByTestId('cirulla-capture-options');
    expect(box.querySelectorAll('button')).toHaveLength(1);
    expect(screen.getByTestId('cirulla-take-2')).toBeInTheDocument();
  });

  it('sends the chosen group with the card', async () => {
    renderWithProviders(<CirullaPage />);
    await pickHand(0);
    fireEvent.click(await screen.findByTestId('cirulla-take-2'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 0, captureIndices: [2] }));
  });

  // **アッソの総取りも 1 つの候補として来る。**
  it('offers the whole table for an ace', async () => {
    renderWithProviders(<CirullaPage />);
    await pickHand(2); // ♦A → [[0,1,2,3]]
    fireEvent.click(await screen.findByTestId('cirulla-take-0-1-2-3'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 2, captureIndices: [0, 1, 2, 3] }));
  });

  // **取れないときだけ置ける。**
  it('offers lay off only when nothing can be captured', async () => {
    renderWithProviders(<CirullaPage />);
    await pickHand(1); // ♥5 → []
    expect(await screen.findByTestId('cirulla-lay-off')).toBeInTheDocument();
    fireEvent.click(screen.getByTestId('cirulla-lay-off'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 1, captureIndices: undefined }));
  });

  it('does not offer lay off while a capture is available', async () => {
    renderWithProviders(<CirullaPage />);
    await pickHand(0);
    await screen.findByTestId('cirulla-take-2');
    expect(screen.queryByTestId('cirulla-lay-off')).not.toBeInTheDocument();
  });

  // **配札ボーナスは出た瞬間に見せる。**
  it('surfaces a deal bonus', async () => {
    const base = makeCirullaState();
    mockExec.mockResolvedValue(
      makeCirullaState({ players: base.players.map((p, i) => (i === 0 ? { ...p, lastBonus: 'barsegon' } : p)) }),
    );
    renderWithProviders(<CirullaPage />);
    expect(await screen.findByTestId('cirulla-bonus-0')).toHaveTextContent('バルセゴン');
    expect(screen.queryByTestId('cirulla-bonus-1')).not.toBeInTheDocument();
  });

  it('breaks the round scoring down and advances', async () => {
    mockExec.mockResolvedValue(
      makeCirullaState({
        phase: 'roundEnd',
        isHumanTurn: false,
        lastResult: {
          lines: [
            { key: 'cards', points: [1, 0] },
            { key: 'denari', points: [0, 0] },
            { key: 'grande', points: [5, 0] },
          ],
          totals: [6, 0],
          sweptDenari: -1,
        },
      }),
    );
    renderWithProviders(<CirullaPage />);
    const box = await screen.findByTestId('cirulla-round-result');
    expect(box).toHaveTextContent('最多枚数');
    expect(box).toHaveTextContent('グランデ');
    // 0 - 0 の項目は並べない。
    expect(box).not.toHaveTextContent('最多デナリ');
    expect(box).toHaveTextContent('ラウンド計: 6 - 0');
    fireEvent.click(screen.getByTestId('cirulla-next-round'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  // **デナリ総取りは即勝ち。** 画面でもそう言う。
  it('calls out a denari sweep', async () => {
    mockExec.mockResolvedValue(
      makeCirullaState({
        phase: 'roundEnd',
        isHumanTurn: false,
        lastResult: { lines: [{ key: 'denari', points: [1, 0] }], totals: [1, 0], sweptDenari: 0 },
      }),
    );
    renderWithProviders(<CirullaPage />);
    expect(await screen.findByTestId('cirulla-swept-denari')).toBeInTheDocument();
  });

  it('names the winner at the end', async () => {
    mockExec.mockResolvedValue(
      makeCirullaState({ phase: 'gameEnd', gameEndFlag: true, isHumanTurn: false, winnerIdx: 1 }),
    );
    renderWithProviders(<CirullaPage />);
    expect(await screen.findByTestId('cirulla-winner')).toHaveTextContent('CPU 1');
  });

  // ヒントのゲート: 頼んでいないヒントは出さない。
  it('does not render the hint banner unless it was requested', async () => {
    mockExec.mockResolvedValue(
      makeCirullaState({ hintHandIdx: 0, hintCaptureIdxs: [2], hintReason: 'capture', messageCode: '' }),
    );
    renderWithProviders(<CirullaPage />);
    await screen.findByTestId('cirulla-table');
    expect(screen.queryByText(/\[0\]/)).not.toBeInTheDocument();
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(
      makeCirullaState({
        hintHandIdx: 0,
        hintCaptureIdxs: [2],
        hintReason: 'capture',
        messageCode: 'cirulla.hintRequested',
      }),
    );
    renderWithProviders(<CirullaPage />);
    expect(await screen.findByText(/\[0\]/)).toBeInTheDocument();
  });
});
