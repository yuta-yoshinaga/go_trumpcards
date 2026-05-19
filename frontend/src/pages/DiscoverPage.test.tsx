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
    const q1 = i18n.t('discover:' + AXES.mood.questionI18nKeys[0]);
    const opt1 = i18n.t('discover:' + AXES.mood.options[0].i18nKey);
    // Sanity: the bundle resolves both keys to real Japanese strings.
    expect(q1).not.toBe('discover:' + AXES.mood.questionI18nKeys[0]);
    expect(opt1).not.toBe('discover:' + AXES.mood.options[0].i18nKey);
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

  it('skip button advances without recording an answer', () => {
    setup();
    // Buttons rendered: 4 mood options + skip; click the last one (skip).
    const buttons = screen.getAllByRole('button');
    const skipBtn = buttons[buttons.length - 1];
    fireEvent.click(skipBtn);
    // After skip, question 2 (skill axis q1) is rendered. The aria-label on
    // the question section names the current question number.
    expect(screen.getByLabelText(/Question 2 of 8/i)).toBeInTheDocument();
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
