import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../test/renderWithProviders';
import type { GongZhuScoreBreakdown } from '../types/games/gongzhu';
import { GongZhuScoreBreakdownPanel } from './GongZhuScoreBreakdownPanel';

const empty: GongZhuScoreBreakdown = {
  heartCount: 0,
  heartsSum: 0,
  allHearts: false,
  aceExposed: false,
  hasPig: false,
  pigExposed: false,
  hasSheep: false,
  sheepExposed: false,
  hasDoubler: false,
  doublerMultiplier: 0,
  doublerStandalone: 0,
  subtotal: 0,
  total: 0,
};

const name = (i: number) => `P${i.toString()}`;

describe('GongZhuScoreBreakdownPanel', () => {
  it('lists every step that happened', () => {
    renderWithProviders(
      <GongZhuScoreBreakdownPanel
        breakdowns={[
          {
            ...empty,
            heartCount: 3,
            heartsSum: -120,
            aceExposed: true,
            hasPig: true,
            pigExposed: true,
            hasSheep: true,
            hasDoubler: true,
            doublerMultiplier: 2,
            subtotal: -220,
            total: -440,
          },
        ]}
        playerName={name}
      />,
    );

    const panel = screen.getByTestId('gz-score-breakdown');
    expect(panel).toHaveTextContent('ハート 3枚: -120');
    expect(panel).toHaveTextContent('♥A 公開により2倍');
    expect(panel).toHaveTextContent('猪 (♠Q) 公開: -200');
    expect(panel).toHaveTextContent('羊 (♦J): +100'); // 非公開のほう
    expect(panel).toHaveTextContent('小計: -220');
    expect(panel).toHaveTextContent('猪抜き (♣10): 2倍');
    expect(panel).toHaveTextContent('合計: -440');
  });

  // **起きていないことは書かない。**取っていない猪の行が出ると、何が起きたのか
  // 読み取れなくなる。
  it('omits the steps that did not happen', () => {
    renderWithProviders(
      <GongZhuScoreBreakdownPanel
        breakdowns={[{ ...empty, heartCount: 2, heartsSum: -90, subtotal: -90, total: -90 }]}
        playerName={name}
      />,
    );

    const panel = screen.getByTestId('gz-score-breakdown');
    expect(panel).not.toHaveTextContent('猪');
    expect(panel).not.toHaveTextContent('羊');
    expect(panel).not.toHaveTextContent('小計');
    expect(panel).toHaveTextContent('合計: -90');
  });

  // 猪抜きが単独で点になるケース (他に得点が無いとき) は倍率ではなく固定点。
  it('shows the doubler as a standalone score when nothing else scored', () => {
    renderWithProviders(
      <GongZhuScoreBreakdownPanel
        breakdowns={[{ ...empty, hasDoubler: true, doublerStandalone: 50, total: 50 }]}
        playerName={name}
      />,
    );

    const panel = screen.getByTestId('gz-score-breakdown');
    expect(panel).toHaveTextContent('猪抜き (♣10) 単独: 50');
    expect(panel).not.toHaveTextContent('倍');
  });
});
