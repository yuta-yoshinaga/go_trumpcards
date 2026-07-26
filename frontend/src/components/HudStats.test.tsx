import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { afTendency, HudStats, overallStyle, pfrTendency, threeBetTendency, vpipTendency } from './HudStats';

describe('HudStats tendency classifiers', () => {
  describe('vpipTendency', () => {
    it.each([
      [10, 'tight'],
      [19, 'tight'],
      [20, 'normal'],
      [25, 'normal'],
      [35, 'normal'],
      [36, 'loose'],
      [60, 'loose'],
    ])('vpip=%d -> %s', (value, want) => {
      expect(vpipTendency(value)).toBe(want);
    });
  });

  describe('pfrTendency', () => {
    it.each([
      [5, 'passive'],
      [9, 'passive'],
      [10, 'balanced'],
      [25, 'balanced'],
      [26, 'aggressive'],
      [50, 'aggressive'],
    ])('pfr=%d -> %s', (value, want) => {
      expect(pfrTendency(value)).toBe(want);
    });
  });

  describe('threeBetTendency', () => {
    it.each([
      [0, 'passive'],
      [3, 'passive'],
      [4, 'balanced'],
      [9, 'balanced'],
      [10, 'aggressive'],
      [20, 'aggressive'],
    ])('threeBet=%d -> %s', (value, want) => {
      expect(threeBetTendency(value)).toBe(want);
    });
  });

  describe('afTendency', () => {
    it.each([
      ['0', 'passive'],
      ['0.9', 'passive'],
      ['1', 'balanced'],
      ['2.5', 'balanced'],
      ['3', 'balanced'],
      ['3.1', 'aggressive'],
      ['10', 'aggressive'],
    ])('af=%s -> %s', (value, want) => {
      expect(afTendency(value)).toBe(want);
    });

    it('non-numeric values fall back to balanced (no misleading red)', () => {
      // The server sometimes ships "∞" or "—" for AF when calls = 0; we must
      // never surface those as "aggressive" since the badge would be a lie.
      expect(afTendency('∞')).toBe('balanced');
      expect(afTendency('—')).toBe('balanced');
      expect(afTendency('')).toBe('balanced');
    });
  });
});

describe('HudStats component', () => {
  it('renders all four telemetry values with tendency markers', () => {
    render(<HudStats vpip={45} pfr={30} threeBet={12} af="3.5" />);
    const hud = screen.getByTestId('hud-stats');
    expect(hud).toBeInTheDocument();
    // Loose VPIP
    expect(screen.getByTestId('hud-vpip-tendency')).toHaveAttribute('data-tendency', 'loose');
    expect(screen.getByTestId('hud-vpip-tendency')).toHaveTextContent('45%');
    // Aggressive PFR / 3Bet / AF
    expect(screen.getByTestId('hud-pfr-tendency')).toHaveAttribute('data-tendency', 'aggressive');
    expect(screen.getByTestId('hud-3bet-tendency')).toHaveAttribute('data-tendency', 'aggressive');
    expect(screen.getByTestId('hud-af-tendency')).toHaveAttribute('data-tendency', 'aggressive');
  });

  it('classifies a tight passive opponent', () => {
    render(<HudStats vpip={15} pfr={5} threeBet={2} af="0.5" />);
    expect(screen.getByTestId('hud-vpip-tendency')).toHaveAttribute('data-tendency', 'tight');
    expect(screen.getByTestId('hud-pfr-tendency')).toHaveAttribute('data-tendency', 'passive');
    expect(screen.getByTestId('hud-3bet-tendency')).toHaveAttribute('data-tendency', 'passive');
    expect(screen.getByTestId('hud-af-tendency')).toHaveAttribute('data-tendency', 'passive');
  });

  it('uses the namespace prop for i18n lookups', () => {
    // The shared component must accept a custom namespace so omaha /
    // shortdeck / pineapple variant pages can render their own tooltip
    // copy without duplicating the component. The default namespace is
    // 'holdem' for backward compatibility.
    render(<HudStats vpip={25} pfr={15} threeBet={6} af="2.0" namespace="omaha" />);
    expect(screen.getByTestId('hud-stats')).toBeInTheDocument();
  });

  it.each([
    ['lag', 45, 30, 'LAG'], // Loose + Aggressive
    ['tag', 15, 30, 'TAG'], // Tight + Aggressive
    ['tp', 15, 5, 'TP'], // Tight + Passive
    ['lp', 45, 5, 'LP'], // Loose + Passive
    ['balanced', 25, 15, 'BAL'], // Normal on either axis → Balanced
  ])('renders the %s style badge for vpip=%d pfr=%d', (expectedStyle, vpip, pfr, expectedLabel) => {
    render(<HudStats vpip={vpip} pfr={pfr} threeBet={6} af="2.0" />);
    const badge = screen.getByTestId('hud-overall-style');
    expect(badge).toHaveAttribute('data-style', expectedStyle);
    expect(badge.textContent).toContain(expectedLabel);
  });

  // Every namespace that a page actually passes must own the `style.*` keys,
  // otherwise i18next falls back to printing the key itself and the badge
  // reads "style.tp" on screen. omahahilo / bigohilo shipped without them.
  // See issue #4374.
  it.each(['holdem', 'omaha', 'shortdeck', 'pineapple', 'bigo', 'omahahilo', 'bigohilo'])(
    'resolves the style badge label in the %s namespace',
    (namespace) => {
      render(<HudStats vpip={15} pfr={5} threeBet={6} af="2.0" namespace={namespace} />);
      const badge = screen.getByTestId('hud-overall-style');
      expect(badge).toHaveAttribute('data-style', 'tp');
      expect(badge.textContent).toContain('TP');
      expect(badge.textContent).not.toContain('style.');
    },
  );
});

describe('overallStyle', () => {
  it('returns TAG for tight + aggressive players', () => {
    expect(overallStyle(15, 30)).toBe('tag');
  });
  it('returns LAG for loose + aggressive players', () => {
    expect(overallStyle(45, 30)).toBe('lag');
  });
  it('returns TP for tight + passive players', () => {
    expect(overallStyle(15, 5)).toBe('tp');
  });
  it('returns LP for loose + passive players', () => {
    expect(overallStyle(45, 5)).toBe('lp');
  });
  it('returns balanced when either dimension is normal', () => {
    expect(overallStyle(25, 15)).toBe('balanced');
    expect(overallStyle(25, 5)).toBe('balanced');
    expect(overallStyle(15, 15)).toBe('balanced');
  });
});
