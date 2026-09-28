import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import type { ReactNode } from 'react';
import { TUTORIAL_NO_SUGGEST_KEY } from '../constants/tutorialKeys';
import { SoundProvider } from '../providers/SoundProvider';

export function renderWithProviders(ui: ReactNode) {
  // Page tests often clear localStorage to isolate game preferences. Reapply
  // the test default here so that this cleanup does not open a real tutorial
  // suggestion dialog over the page under test.
  localStorage.setItem(TUTORIAL_NO_SUGGEST_KEY, 'true');
  const queryClient = new QueryClient({
    defaultOptions: {
      mutations: { retry: false },
    },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <SoundProvider>{ui}</SoundProvider>
    </QueryClientProvider>,
  );
}
