import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { HashRouter, Route, Routes } from 'react-router-dom';
import { ErrorBoundary } from './components/ErrorBoundary';
import { NavBar } from './components/NavBar';
import { SkipNavLink } from './components/SkipNavLink';
import { gameRoutes } from './constants/gameRoutes';
import { BaccaratPage } from './pages/BaccaratPage';
import { BlackJackPage } from './pages/BlackJackPage';
import { CrazyEightsPage } from './pages/CrazyEightsPage';
import { DaifugoPage } from './pages/DaifugoPage';
import { DoubtPage } from './pages/DoubtPage';
import { FreeCellPage } from './pages/FreeCellPage';
import { GinRummyPage } from './pages/GinRummyPage';
import { HeartsPage } from './pages/HeartsPage';
import { HoldemPage } from './pages/HoldemPage';
import { KlondikePage } from './pages/KlondikePage';
import { MemoryPage } from './pages/MemoryPage';
import { OldMaidPage } from './pages/OldMaidPage';
import { OmahaPage } from './pages/OmahaPage';
import { PokerPage } from './pages/PokerPage';
import { SevensPage } from './pages/SevensPage';
import { SpadesPage } from './pages/SpadesPage';

type GamePath = (typeof gameRoutes)[number]['path'];
const pageByPath: Record<GamePath, ReactNode> = {
  '/': <BlackJackPage />,
  '/poker': <PokerPage />,
  '/oldmaid': <OldMaidPage />,
  '/daifugo': <DaifugoPage />,
  '/sevens': <SevensPage />,
  '/doubt': <DoubtPage />,
  '/holdem': <HoldemPage />,
  '/omaha': <OmahaPage />,
  '/hearts': <HeartsPage />,
  '/spades': <SpadesPage />,
  '/memory': <MemoryPage />,
  '/klondike': <KlondikePage />,
  '/freecell': <FreeCellPage />,
  '/baccarat': <BaccaratPage />,
  '/crazyeights': <CrazyEightsPage />,
  '/ginrummy': <GinRummyPage />,
};

/** Root application component with router and game page routes. */
export default function App() {
  const { t } = useTranslation();
  return (
    <HashRouter>
      <ErrorBoundary>
        <div className="flex flex-col h-full">
          <SkipNavLink targetId="main-content" label={t('nav.skipToContent')} />
          <NavBar />
          <main id="main-content" className="flex-1 flex flex-col min-h-0">
            <Routes>
              {gameRoutes.map(({ path }) => (
                <Route key={path} path={path} element={pageByPath[path]} />
              ))}
            </Routes>
          </main>
        </div>
      </ErrorBoundary>
    </HashRouter>
  );
}
