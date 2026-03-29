import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { HashRouter, Route, Routes } from 'react-router-dom';
import { DesktopSidebar } from './components/DesktopSidebar';
import { ErrorBoundary } from './components/ErrorBoundary';
import { NavBar } from './components/NavBar';
import { SkipNavLink } from './components/SkipNavLink';
import { gameRoutes } from './constants/gameRoutes';
import { useGameSound } from './hooks/useGameSound';
import { BaccaratPage } from './pages/BaccaratPage';
import { BlackJackPage } from './pages/BlackJackPage';
import { CrazyEightsPage } from './pages/CrazyEightsPage';
import { CribbagePage } from './pages/CribbagePage';
import { DaifugoPage } from './pages/DaifugoPage';
import { DeucesWildPage } from './pages/DeucesWildPage';
import { DoubtPage } from './pages/DoubtPage';
import { EuchrePage } from './pages/EuchrePage';
import { FreeCellPage } from './pages/FreeCellPage';
import { GinRummyPage } from './pages/GinRummyPage';
import { HeartsPage } from './pages/HeartsPage';
import { HoldemPage } from './pages/HoldemPage';
import { IndianPokerPage } from './pages/IndianPokerPage';
import { JokerPokerPage } from './pages/JokerPokerPage';
import { KlondikePage } from './pages/KlondikePage';
import { MemoryPage } from './pages/MemoryPage';
import { NapoleonPage } from './pages/NapoleonPage';
import { OldMaidPage } from './pages/OldMaidPage';
import { OmahaPage } from './pages/OmahaPage';
import { PokerPage } from './pages/PokerPage';
import { PyramidPage } from './pages/PyramidPage';
import { SevensPage } from './pages/SevensPage';
import { ShortDeckPage } from './pages/ShortDeckPage';
import { SpadesPage } from './pages/SpadesPage';
import { SpiderPage } from './pages/SpiderPage';
import { ThreeCardPage } from './pages/ThreeCardPage';
import { TriPeaksPage } from './pages/TriPeaksPage';
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
  '/shortdeck': <ShortDeckPage />,
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
  '/tripeaks': <TriPeaksPage />,
  '/indianpoker': <IndianPokerPage />,
  '/videopoker': <VideoPokerPage />,
  '/deuceswild': <DeucesWildPage />,
  '/jokerpoker': <JokerPokerPage />,
  '/threecard': <ThreeCardPage />,
};

/** Root application component with router and game page routes. */
export default function App() {
  const { t } = useTranslation();
  const { muted, toggleMute } = useGameSound();
  return (
    <HashRouter>
      <ErrorBoundary>
        <div className="flex flex-col h-full lg:flex-row">
          <SkipNavLink targetId="main-content" label={t('nav.skipToContent')} />
          <DesktopSidebar soundMuted={muted} onSoundToggle={toggleMute} />
          <div className="flex flex-col flex-1 min-w-0">
            <NavBar soundMuted={muted} onSoundToggle={toggleMute} />
            <main id="main-content" tabIndex={-1} className="flex-1 flex flex-col min-h-0">
              <Routes>
                {gameRoutes.map(({ path }) => (
                  <Route key={path} path={path} element={pageByPath[path]} />
                ))}
              </Routes>
            </main>
          </div>
        </div>
      </ErrorBoundary>
    </HashRouter>
  );
}
