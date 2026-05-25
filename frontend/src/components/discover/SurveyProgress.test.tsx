/**
 * @vitest-environment jsdom
 */
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { TOTAL_QUESTIONS } from '../../constants/discoverAxes';
import { SurveyProgress } from './SurveyProgress';

describe('SurveyProgress', () => {
  it('renders TOTAL_QUESTIONS cards', () => {
    render(<SurveyProgress current={1} />);
    expect(screen.getAllByTestId('survey-progress-card')).toHaveLength(TOTAL_QUESTIONS);
  });

  it('marks exactly one card as current', () => {
    render(<SurveyProgress current={3} />);
    const cards = screen.getAllByTestId('survey-progress-card');
    const current = cards.filter((c) => c.dataset.state === 'current');
    expect(current).toHaveLength(1);
    expect(cards.indexOf(current[0])).toBe(2);
  });

  it('marks previous cards as done', () => {
    render(<SurveyProgress current={5} />);
    const cards = screen.getAllByTestId('survey-progress-card');
    const done = cards.filter((c) => c.dataset.state === 'done');
    expect(done).toHaveLength(4);
  });

  it('announces a "started" message on first mount (#1903), not "moved to 1 of 8"', () => {
    // Regression: ISSUE-008 — `aria.progress` fired on initial mount,
    // saying "Moved to 1 / 8 に進みました" / "Moved to 1 of 8" — which
    // is wrong copy because the user didn't *advance*, they arrived.
    // Found by /qa on 2026-05-20.
    const { container } = render(<SurveyProgress current={1} />);
    const live = container.querySelector('[aria-live="polite"]');
    expect(live).not.toBeNull();
    expect(live?.textContent ?? '').not.toMatch(/進みました|Moved to/);
  });

  it('announces a "moved to" message when current > 1', () => {
    const { container } = render(<SurveyProgress current={3} />);
    const live = container.querySelector('[aria-live="polite"]');
    expect(live?.textContent ?? '').toMatch(/進みました/);
  });
});
