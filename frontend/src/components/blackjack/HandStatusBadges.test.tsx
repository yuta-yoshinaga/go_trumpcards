import { render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { describe, expect, it } from 'vitest';
import { HandStatusBadges } from './HandStatusBadges';

const noBadges = { busted: false, doubled: false, isBlackJack: false, surrendered: false };

describe('HandStatusBadges', () => {
  it('renders nothing when all flags are false', () => {
    const { container } = render(<HandStatusBadges {...noBadges} />);
    expect(container.querySelectorAll('abbr')).toHaveLength(0);
  });

  it('renders BUST badge with tooltip when busted is true', () => {
    render(<HandStatusBadges {...noBadges} busted={true} />);
    const elem = screen.getByTitle(i18n.t('blackjack:status.bustTooltip'));
    expect(elem).toBeInTheDocument();
    expect(elem).toHaveTextContent(`[${i18n.t('blackjack:status.bust')}]`);
  });

  it('renders DD badge with tooltip when doubled is true', () => {
    render(<HandStatusBadges {...noBadges} doubled={true} />);
    const elem = screen.getByTitle(i18n.t('blackjack:status.ddTooltip'));
    expect(elem).toBeInTheDocument();
    expect(elem).toHaveTextContent(`[${i18n.t('blackjack:status.dd')}]`);
  });

  it('renders BJ badge with tooltip when isBlackJack is true', () => {
    render(<HandStatusBadges {...noBadges} isBlackJack={true} />);
    const elem = screen.getByTitle(i18n.t('blackjack:status.bjTooltip'));
    expect(elem).toBeInTheDocument();
    expect(elem).toHaveTextContent(`[${i18n.t('blackjack:status.bj')}]`);
  });

  it('renders SUR badge with pill styling and tooltip when surrendered is true', () => {
    render(<HandStatusBadges {...noBadges} surrendered={true} />);
    const elem = screen.getByTitle(i18n.t('blackjack:status.surTooltip'));
    expect(elem).toBeInTheDocument();
    expect(elem).toHaveTextContent(`[${i18n.t('blackjack:status.sur')}]`);
    expect(elem).toHaveClass('bg-ds-surface-elevated');
  });

  it('renders all badges simultaneously when all flags are true', () => {
    render(<HandStatusBadges busted={true} doubled={true} isBlackJack={true} surrendered={true} />);
    expect(screen.getByTitle(i18n.t('blackjack:status.bustTooltip'))).toBeInTheDocument();
    expect(screen.getByTitle(i18n.t('blackjack:status.ddTooltip'))).toBeInTheDocument();
    expect(screen.getByTitle(i18n.t('blackjack:status.bjTooltip'))).toBeInTheDocument();
    expect(screen.getByTitle(i18n.t('blackjack:status.surTooltip'))).toBeInTheDocument();
  });
});
