import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import i18n from '../../i18n';
import { PokerStatsBadge } from './PokerStatsBadge';

const t = i18n.getFixedT(null, 'fivecardstud');

const stats = { totalHands: 40, vpip: 25, pfr: 12, threeBet: 5, af: '2.3' };

describe('PokerStatsBadge', () => {
  it('renders every statistic the CUI reports', () => {
    render(<PokerStatsBadge stats={stats} t={t} />);
    const badge = screen.getByTestId('fcs-stats');
    for (const fragment of ['25', '12', '5', '2.3']) {
      expect(badge).toHaveTextContent(fragment);
    }
  });

  // **未プレイでは出さない。**統計として意味が無く、CUI も同じ条件で行ごと
  // 省いている。ここを描画してしまうと「VPIP 0%」が実績のように読める。
  it('renders nothing before any hand has been played', () => {
    const { container } = render(<PokerStatsBadge stats={{ ...stats, totalHands: 0 }} t={t} />);
    expect(container).toBeEmptyDOMElement();
  });

  // AF はバックエンドが整形済みの文字列で、割り算不能なとき "∞" が来る。
  // 数値として扱うと NaN になるため、そのまま出せることを踏む。
  it('passes the pre-formatted aggression factor through unchanged', () => {
    render(<PokerStatsBadge stats={{ ...stats, af: '∞' }} t={t} />);
    expect(screen.getByTestId('fcs-stats')).toHaveTextContent('∞');
  });
});
