/**
 * @vitest-environment jsdom
 */
import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { AXES } from '../constants/discoverAxes';
import { renderWithProviders } from '../test/renderWithProviders';
import { DiscoverPage } from './DiscoverPage';

function LocationProbe({ onPath }: { onPath: (p: string) => void }) {
  const loc = useLocation();
  onPath(`${loc.pathname}${loc.search}`);
  return null;
}

function setup(initialEntries: string[] = ['/discover']) {
  let lastPath = initialEntries[0];
  renderWithProviders(
    <MemoryRouter initialEntries={initialEntries}>
      <Routes>
        <Route path="/discover" element={<DiscoverPage />} />
        <Route
          path="/discover/result"
          element={
            <LocationProbe
              onPath={(p) => {
                lastPath = p;
              }}
            />
          }
        />
      </Routes>
    </MemoryRouter>,
  );
  return { getLastPath: () => lastPath };
}

describe('DiscoverPage', () => {
  beforeEach(() => {
    localStorage.clear();
  });
  afterEach(() => {
    localStorage.clear();
  });

  it('renders the first question on mount', () => {
    setup();
    expect(screen.getByRole('heading')).toBeInTheDocument();
  });

  it('renders translated question and option text, not raw i18n keys', () => {
    setup();
    const q1Key = AXES.mood.questions[0].questionI18nKey;
    const opt1Key = AXES.mood.questions[0].options[0].i18nKey;
    const q1 = i18n.t(`discover:${q1Key}`);
    const opt1 = i18n.t(`discover:${opt1Key}`);
    // Sanity: the bundle resolves both keys to real Japanese strings.
    expect(q1).not.toBe(`discover:${q1Key}`);
    expect(opt1).not.toBe(`discover:${opt1Key}`);
    expect(screen.getByRole('heading', { level: 2, name: q1 })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: new RegExp(opt1) })).toBeInTheDocument();
  });

  it('advances through 8 questions via keyboard digit shortcuts and navigates to /discover/result', async () => {
    const { getLastPath } = setup();
    // Each answer triggers a re-render which re-registers the listener with
    // the new `current` — we need to await each iteration so the next key
    // event sees the updated handler closure.
    for (let i = 0; i < 8; i++) {
      await act(async () => {
        fireEvent.keyDown(window, { key: '1' });
      });
    }
    await waitFor(() => {
      expect(getLastPath()).toMatch(/^\/discover\/result\?/);
    });
  });

  it('submits the actual answer values to the URL, not all blanks (#1896)', async () => {
    // Regression: ISSUE-001 — the submit useEffect listed `axes` in its dep
    // array and called resetDraft() which mutated `axes` to EMPTY_ANSWERS.
    // That fired the effect a second time, re-encoding the now-empty draft
    // into the URL and silently overwriting the just-pushed good URL.
    // Found by /qa on 2026-05-20
    // Report: .gstack/qa-reports/qa-report-go-trumpcards-dev-2026-05-20.md
    const { getLastPath } = setup();
    for (let i = 0; i < 8; i++) {
      await act(async () => {
        fireEvent.keyDown(window, { key: '1' });
      });
    }
    await waitFor(() => {
      expect(getLastPath()).toMatch(/^\/discover\/result\?/);
    });
    const finalPath = getLastPath();
    // Every key '1' selects option index 0 — every axis slot must be '0,0'.
    // Pre-fix, this was 'm=-,-&s=-,-&so=-,-&t=-,-' regardless of the answers.
    expect(finalPath).toContain('m=0,0');
    expect(finalPath).toContain('s=0,0');
    expect(finalPath).toContain('so=0,0');
    expect(finalPath).toContain('t=0,0');
    expect(finalPath).not.toMatch(/=-,-/);
  });

  it('skip button advances without recording an answer', () => {
    setup();
    // Buttons rendered: mood Q1 options + skip; click the last one (skip).
    const buttons = screen.getAllByRole('button');
    const skipBtn = buttons[buttons.length - 1];
    fireEvent.click(skipBtn);
    // After skip, question 2 (skill axis q1) is rendered. The aria-label on
    // the question section names the current question number.
    expect(screen.getByLabelText(/Question 2 of 8/i)).toBeInTheDocument();
  });

  it('ignores rapid double-fire of the same option (#1898)', async () => {
    // Regression: ISSUE-003 — two rapid clicks on an option used to fire
    // handleSelect twice with the same captured `current`, calling
    // dispatch({type:'advance'}) twice and skipping the next question. The
    // user would land on Q3 instead of Q2, with the in-between sub-dim
    // silently left null. Verified by /qa on 2026-05-20.
    setup();
    await act(async () => {
      // Both fires happen inside one batch — both see the Q1 closure.
      fireEvent.keyDown(window, { key: '1' });
      fireEvent.keyDown(window, { key: '1' });
    });
    // After a single answer, we expect Q2 — not Q3.
    expect(screen.getByLabelText(/Question 2 of 8/i)).toBeInTheDocument();
  });

  it('browser back walks to the previous question instead of exiting Discover (#1899)', async () => {
    // Regression: ISSUE-004 — the browser back button used to pop /discover
    // off the history stack entirely, dumping the user back on the previous
    // page. With the popstate listener and per-advance pushState, back walks
    // questions inside the survey. Found by /qa on 2026-05-20.
    setup();
    // Two advances cover both the forward-push effect AND the subsequent
    // backward branch (step > 0 with direction === 'backward' skips the push).
    await act(async () => {
      fireEvent.keyDown(window, { key: '1' });
    });
    await act(async () => {
      fireEvent.keyDown(window, { key: '1' });
    });
    expect(screen.getByLabelText(/Question 3 of 8/i)).toBeInTheDocument();
    // popstate walks back to Q2.
    await act(async () => {
      window.dispatchEvent(new PopStateEvent('popstate'));
    });
    expect(screen.getByLabelText(/Question 2 of 8/i)).toBeInTheDocument();
    // popstate again walks back to Q1.
    await act(async () => {
      window.dispatchEvent(new PopStateEvent('popstate'));
    });
    expect(screen.getByLabelText(/Question 1 of 8/i)).toBeInTheDocument();
  });

  it('popstate at the first question is a no-op (no exit, no negative step) (#1899)', async () => {
    setup();
    // Mount at Q1 — fire popstate without advancing. Handler must not
    // dispatch since state.step === 0, and the user stays on Q1.
    await act(async () => {
      window.dispatchEvent(new PopStateEvent('popstate'));
    });
    expect(screen.getByLabelText(/Question 1 of 8/i)).toBeInTheDocument();
  });

  it('Backspace returns to the previous question', () => {
    setup();
    act(() => {
      fireEvent.keyDown(window, { key: '1' });
    });
    // Now on question 2.
    expect(screen.getByLabelText(/Question 2 of 8/i)).toBeInTheDocument();
    act(() => {
      fireEvent.keyDown(window, { key: 'Backspace' });
    });
    // Back to question 1.
    expect(screen.getByLabelText(/Question 1 of 8/i)).toBeInTheDocument();
  });
});
