/**
 * @vitest-environment jsdom
 */
import { render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { gameRoutes } from '../../constants/gameRoutes';
import { StretchPickCard } from './StretchPickCard';

const briscola = gameRoutes.find((r) => r.path === '/briscola');
if (!briscola) throw new Error('test fixture: Briscola route missing');

describe('StretchPickCard', () => {
  it('renders the stretch eyebrow + a link to the game', () => {
    render(
      <MemoryRouter>
        <StretchPickCard game={briscola} />
      </MemoryRouter>,
    );
    expect(screen.getByRole('heading')).toBeInTheDocument();
    expect(screen.getByRole('link')).toHaveAttribute('href', briscola.path);
  });

  it('renders the blurb fallback (not the raw i18n key) when stretch.<slug> is missing (#1897)', () => {
    // Regression: ISSUE-002 — `stretch.<slug>` entries were never added to
    // discover.json, so every Stretch card showed the literal key string
    // (e.g. "stretch.briscola") in place of the description.
    // Found by /qa on 2026-05-20
    // Report: .gstack/qa-reports/qa-report-go-trumpcards-dev-2026-05-20.md
    const slug = briscola.page.toLowerCase();
    const blurb = i18n.t(`discover:blurb.${slug}`);
    // Sanity: the blurb namespace really has briscola.
    expect(blurb).not.toBe(`discover:blurb.${slug}`);
    expect(blurb).not.toBe(`blurb.${slug}`);

    render(
      <MemoryRouter>
        <StretchPickCard game={briscola} />
      </MemoryRouter>,
    );
    // The raw stretch key must never reach the DOM.
    expect(screen.queryByText(`stretch.${slug}`)).toBeNull();
    // The blurb fallback should be visible.
    expect(screen.getByText(blurb)).toBeInTheDocument();
  });
});
