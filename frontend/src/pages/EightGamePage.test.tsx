import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { eightGameApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeHorseState } from '../test/stateFactories';
import { EightGamePage } from './EightGamePage';

vi.mock('../api/gameApi', () => ({
  eightGameApi: { exec: vi.fn() },
  horseApi: { exec: vi.fn() },
  actionLogApi: { eightgame: vi.fn() },
}));

const mockExec = vi.mocked(eightGameApi.exec);

/** A four-handed Eight-Game table, by default in the hold'em discipline. */
const eightGameState = makeHorseState({ variant: 1, rotation: [0, 1, 2, 3, 4, 5, 6, 7] });

/** The same table during the second draw of a 2-7 Triple Draw hand. */
const drawState = makeHorseState({
  variant: 1,
  rotation: [0, 1, 2, 3, 4, 5, 6, 7],
  discipline: 7,
  disciplineLetter: '2-7',
  disciplineName: 'tripleDraw',
  tablePhase: 3,
  isDrawPhase: true,
  drawIndex: 2,
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(eightGameState);
});

describe('EightGamePage', () => {
  // **叩く先が違う。** 同じページを 2 つのゲームが共有しているので、
  // ここが horse のままだと Eight-Game Mix の卓に手が届かない。
  it('drives the eightgame endpoint, not the horse one', async () => {
    renderWithProviders(<EightGamePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('names the running discipline', async () => {
    renderWithProviders(<EightGamePage />);
    const info = await screen.findByTestId('ho-discipline');
    expect(info).toHaveTextContent('テキサスホールデム');
  });

  // **8 種目のほうは 4 人卓しか作れない。** 6 を選べると、6 種目目で理由も
  // 出さずにマッチが終わる卓が作れてしまう。
  it('offers four seats only', async () => {
    renderWithProviders(<EightGamePage />);
    const select = (await screen.findByLabelText('席数')) as HTMLSelectElement;
    expect([...select.options].map((o) => o.value)).toEqual(['4']);
  });

  // **引き直しの番はベットの番ではない。** 両方出すと、賭ける手と引く手が
  // 同時に並ぶ。
  it('swaps the betting controls for the draw controls', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<EightGamePage />);
    expect(await screen.findByTestId('ho-draw')).toHaveTextContent('引き直し 2 回目');
    expect(screen.queryByRole('button', { name: 'コール' })).not.toBeInTheDocument();
  });

  it('keeps the betting controls when no draw is pending', async () => {
    renderWithProviders(<EightGamePage />);
    expect(await screen.findByRole('button', { name: 'コール' })).toBeInTheDocument();
    expect(screen.queryByTestId('ho-draw')).not.toBeInTheDocument();
  });

  // **選んだ札だけを 0 始まりで送る。** 1 始まりで送ると 1 枚ずれた札が捨てられ、
  // 画面には「交換した」としか出ない。
  it('sends the selected card indices, zero-based and sorted', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<EightGamePage />);
    fireEvent.click(await screen.findByTestId('ho-draw-card-1'));
    fireEvent.click(screen.getByTestId('ho-draw-card-0'));
    fireEvent.click(screen.getByTestId('ho-draw-exchange'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw', { cardIndices: [0, 1] }));
  });

  // **何も選んでいないうちは交換できない。** 空で送るとスタンドパットとして
  // 通ってしまい、押し間違いが「引かない」に化ける。
  it('disables the exchange button until a card is picked', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<EightGamePage />);
    expect(await screen.findByTestId('ho-draw-exchange')).toBeDisabled();
    fireEvent.click(screen.getByTestId('ho-draw-card-0'));
    expect(screen.getByTestId('ho-draw-exchange')).toBeEnabled();
  });

  // スタンドパットは空で送る。「引かない」も 1 つの手。
  it('stands pat with an empty index list', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<EightGamePage />);
    fireEvent.click(await screen.findByTestId('ho-draw-stand'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw', { cardIndices: [] }));
  });

  it('names the draw round instead of a street', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<EightGamePage />);
    expect(await screen.findByTestId('ho-round')).toHaveTextContent('引き直し');
  });
});
