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
import { CribbagePage } from './pages/CribbagePage';
import { DaifugoPage } from './pages/DaifugoPage';
import { DoubtPage } from './pages/DoubtPage';
import { EuchrePage } from './pages/EuchrePage';
import { FreeCellPage } from './pages/FreeCellPage';
import { GinRummyPage } from './pages/GinRummyPage';
import { HeartsPage } from './pages/HeartsPage';
import { HoldemPage } from './pages/HoldemPage';
import { IndianPokerPage } from './pages/IndianPokerPage';
import { KlondikePage } from './pages/KlondikePage';
import { MemoryPage } from './pages/MemoryPage';
import { NapoleonPage } from './pages/NapoleonPage';
import { OldMaidPage } from './pages/OldMaidPage';
import { OmahaPage } from './pages/OmahaPage';
import { PokerPage } from './pages/PokerPage';
import { PyramidPage } from './pages/PyramidPage';
import { SevensPage } from './pages/SevensPage';
import { SpadesPage } from './pages/SpadesPage';
import { SpiderPage } from './pages/SpiderPage';
import { VideoPokerPage } from './pages/VideoPokerPage';

type GamePath = (typeof gameRoutes)[number]['path'];
const pageByPath: Record<GamePath, ReactNode> = {
  '/': <BlackJackPage />,
  '/poker': <PokerPage />,
  '/oldmaid': <OldMaidPage />,
  '/daifugo': <DaifugoPage />,
  '/sevens': <SevensPage />,
  '/doubt': <DoubtPage />,
  '/euchre': <EuchrePage />,
  '/holdem': <HoldemPage />,
  '/omaha': <OmahaPage />,
  '/hearts': <HeartsPage />,
  '/spades': <SpadesPage />,
  '/napoleon': <NapoleonPage />,
  '/memory': <MemoryPage />,
  '/klondike': <KlondikePage />,
  '/freecell': <FreeCellPage />,
  '/baccarat': <BaccaratPage />,
  '/crazyeights': <CrazyEightsPage />,
  '/ginrummy': <GinRummyPage />,
  '/cribbage': <CribbagePage />,
  '/spider': <SpiderPage />,
  '/pyramid': <PyramidPage />,
  '/indianpoker': <IndianPokerPage />,
  '/videopoker': <VideoPokerPage />,
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
          <main id="main-content" tabIndex={-1} className="flex-1 flex flex-col min-h-0">
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
