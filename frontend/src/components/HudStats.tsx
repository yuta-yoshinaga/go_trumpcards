import { useId } from 'react';
import { useTranslation } from 'react-i18next';

/**
 * Tendency labels classify a player's playing style from their HUD stats so
 * a non-poker-savvy user can read "VPIP 45% (Loose)" instead of having to
 * memorize that 45% is a wide range. The thresholds match commonly-cited
 * poker theory ranges (Tight < 20 < Normal < 35 < Loose; Passive < 1 AF < 3
 * < Aggressive). They are not meant to be precise — the goal is intuitive
 * color-coded triage so casual players can enjoy the CPU style variation
 * (TAG / LAG / GTO etc.) without studying definitions first. See issue #1564.
 */
type Tendency = 'tight' | 'normal' | 'loose' | 'passive' | 'balanced' | 'aggressive';

/**
 * StatTooltip renders a label that reveals a tooltip on hover/focus.
 * Mirrors the inline tooltips that previously lived on each poker page;
 * extracted to deduplicate four near-identical copies. Uses `useId` so
 * each instance gets a unique DOM id even when multiple HudStats rows
 * are rendered (one per CPU player). See issue #1564.
 */
function StatTooltip({ label, tooltipText }: { label: string; tooltipText: string }) {
  const id = useId();
  return (
    <button
      type="button"
      className="group relative cursor-help bg-transparent border-none p-0 font-inherit text-inherit inline"
      aria-describedby={id}
    >
      {label}
      <span
        id={id}
        className="pointer-events-none absolute bottom-full left-1/2 -translate-x-1/2 mb-1 whitespace-nowrap rounded bg-ds-surface-elevated px-2 py-1 text-xs text-ds-text-primary opacity-0 group-hover:opacity-100 group-focus-within:opacity-100"
        role="tooltip"
      >
        {tooltipText}
      </span>
    </button>
  );
}

/**
 * vpipTendency classifies VPIP into Tight / Normal / Loose.
 * Exported for direct test access — the thresholds are the SSoT for both
 * the rendered badge and any future linter rules. See issue #1564.
 */
export function vpipTendency(value: number): Tendency {
  if (value > 35) return 'loose';
  if (value < 20) return 'tight';
  return 'normal';
}

/**
 * pfrTendency classifies PFR (preflop raise %) into Passive / Balanced /
 * Aggressive. PFR strictly less than VPIP is the norm; the absolute value
 * is what drives perceived aggression. See issue #1564.
 */
export function pfrTendency(value: number): Tendency {
  if (value > 25) return 'aggressive';
  if (value < 10) return 'passive';
  return 'balanced';
}

/**
 * threeBetTendency classifies 3Bet % into Passive / Balanced / Aggressive.
 * See issue #1564.
 */
export function threeBetTendency(value: number): Tendency {
  if (value > 9) return 'aggressive';
  if (value < 4) return 'passive';
  return 'balanced';
}

/**
 * afTendency classifies AF (aggression factor) into Passive / Balanced /
 * Aggressive. AF arrives as a string ("1.5", "∞", "0") because the server
 * formats it; non-numeric values fall back to "balanced" so a weird
 * payload cannot surface a misleading red badge. See issue #1564.
 */
export function afTendency(af: string): Tendency {
  const n = Number.parseFloat(af);
  if (!Number.isFinite(n)) return 'balanced';
  if (n > 3) return 'aggressive';
  if (n < 1) return 'passive';
  return 'balanced';
}

/** Combined VPIP+PFR profile that summarises a CPU's overall poker style. */
export type PokerStyle = 'tag' | 'lag' | 'tap' | 'lap' | 'balanced';

/**
 * overallStyle classifies a player into one of the canonical poker styles
 * from the combination of VPIP (looseness) and PFR (aggression), so the
 * UI can surface a single "TAG / LAG / TP / LP" tag instead of forcing
 * the player to read four percentages and combine them themselves.
 * See issue #1873.
 */
export function overallStyle(vpip: number, pfr: number): PokerStyle {
  const loose = vpipTendency(vpip) === 'loose';
  const tight = vpipTendency(vpip) === 'tight';
  const aggressive = pfrTendency(pfr) === 'aggressive';
  const passive = pfrTendency(pfr) === 'passive';
  if (tight && aggressive) return 'tag';
  if (loose && aggressive) return 'lag';
  if (tight && passive) return 'tap';
  if (loose && passive) return 'lap';
  return 'balanced';
}

const STYLE_ICONS: Record<PokerStyle, string> = {
  tag: '🛡️',
  lag: '⚔️',
  tap: '🐢',
  lap: '🐠',
  balanced: '⚖️',
};

/**
 * tendencyClass maps a tendency to the design-system color class used for
 * the badge text. Loose / Aggressive surface as `text-ds-error` (warm
 * warning red — high engagement), Tight / Passive as `text-ds-info` (cool
 * blue — calm/conservative), and Normal / Balanced as `text-ds-text-muted`
 * so they recede visually. See issue #1564.
 */
function tendencyClass(t: Tendency): string {
  switch (t) {
    case 'loose':
    case 'aggressive':
      return 'text-ds-error';
    case 'tight':
    case 'passive':
      return 'text-ds-info';
    default:
      return 'text-ds-text-muted';
  }
}

/**
 * HudStats props — accepts the four poker telemetry fields that ship in
 * every poker variant's API response, plus an optional i18n namespace
 * that lets variant pages (e.g. crazypineapple sharing logic with
 * pineapple) reuse this component without duplicating tooltip copy.
 */
type HudStatsProps = {
  vpip: number;
  pfr: number;
  threeBet: number;
  af: string;
  /** i18n namespace that owns `stats.*` and `tendency.*` keys. Defaults
   * to "holdem" so existing callers continue to work. Variants pass
   * their own ("omaha", "shortdeck", "pineapple", …). */
  namespace?: string;
};

/**
 * HudStats renders the per-CPU poker telemetry row (VPIP / PFR / 3Bet /
 * AF) with color-coded tendency badges (Tight/Loose, Passive/Aggressive)
 * so casual players can read CPU style at a glance instead of memorizing
 * threshold percentages. Replaces the four near-identical local copies
 * that previously lived on Holdem/Omaha/ShortDeck/Pineapple pages. See
 * issue #1564.
 */
export function HudStats({ vpip, pfr, threeBet, af, namespace = 'holdem' }: HudStatsProps) {
  const { t } = useTranslation(namespace);
  const vpipT = vpipTendency(vpip);
  const pfrT = pfrTendency(pfr);
  const threeBetT = threeBetTendency(threeBet);
  const afT = afTendency(af);
  const style = overallStyle(vpip, pfr);
  return (
    <span className="ml-2 text-ds-info text-[0.8em] hidden md:inline" data-testid="hud-stats">
      <span
        data-testid="hud-overall-style"
        data-style={style}
        title={t(`style.${style}Tooltip`)}
        className="mr-1 inline-flex items-center gap-0.5 px-1 py-0.5 rounded bg-black/30 text-ds-text-primary"
      >
        <span aria-hidden="true">{STYLE_ICONS[style]}</span>
        <span className="font-bold uppercase">{t(`style.${style}`)}</span>
      </span>
      <StatTooltip label={t('stats.vpip')} tooltipText={t('stats.vpipTooltip')} />:
      <span className={tendencyClass(vpipT)} data-testid="hud-vpip-tendency" data-tendency={vpipT}>
        {vpip}%
      </span>{' '}
      <StatTooltip label={t('stats.pfr')} tooltipText={t('stats.pfrTooltip')} />:
      <span className={tendencyClass(pfrT)} data-testid="hud-pfr-tendency" data-tendency={pfrT}>
        {pfr}%
      </span>{' '}
      <StatTooltip label={t('stats.threeBet')} tooltipText={t('stats.threeBetTooltip')} />:
      <span className={tendencyClass(threeBetT)} data-testid="hud-3bet-tendency" data-tendency={threeBetT}>
        {threeBet}%
      </span>{' '}
      <StatTooltip label={t('stats.af')} tooltipText={t('stats.afTooltip')} />:
      <span className={tendencyClass(afT)} data-testid="hud-af-tendency" data-tendency={afT}>
        {af}
      </span>
    </span>
  );
}
