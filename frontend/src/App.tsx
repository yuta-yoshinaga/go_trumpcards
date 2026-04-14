import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { HashRouter, Route, Routes } from 'react-router-dom';
import { DesktopSidebar } from './components/DesktopSidebar';
import { ErrorBoundary } from './components/ErrorBoundary';
import { NavBar } from './components/NavBar';
import { SkipNavLink } from './components/SkipNavLink';
import { gameRoutes } from './constants/gameRoutes';
import { BaccaratPage } from './pages/BaccaratPage';
import { BlackJackPage } from './pages/BlackJackPage';
import { BridgePage } from './pages/BridgePage';
import { CanastaPage } from './pages/CanastaPage';
import { CanfieldPage } from './pages/CanfieldPage';
import { CaribbeanStudPage } from './pages/CaribbeanStudPage';
import { ClockSolitairePage } from './pages/ClockSolitairePage';
import { CrazyEightsPage } from './pages/CrazyEightsPage';
import { CribbagePage } from './pages/CribbagePage';
import { DaifugoPage } from './pages/DaifugoPage';
import { DeucesWildPage } from './pages/DeucesWildPage';
import { DoubtPage } from './pages/DoubtPage';
import { DurakPage } from './pages/DurakPage';
import { EuchrePage } from './pages/EuchrePage';
import { FiftyOnePage } from './pages/FiftyOnePage';
import { FortyThievesPage } from './pages/FortyThievesPage';
import { FreeCellPage } from './pages/FreeCellPage';
import { GinRummyPage } from './pages/GinRummyPage';
import { GoFishPage } from './pages/GoFishPage';
import { GolfPage } from './pages/GolfPage';
import { HeartsPage } from './pages/HeartsPage';
import { HoldemPage } from './pages/HoldemPage';
import { IndianPokerPage } from './pages/IndianPokerPage';
import { JokerPokerPage } from './pages/JokerPokerPage';
import { KlondikePage } from './pages/KlondikePage';
import { LetItRidePage } from './pages/LetItRidePage';
import { MemoryPage } from './pages/MemoryPage';
import { NapoleonPage } from './pages/NapoleonPage';
import { OhHellPage } from './pages/OhHellPage';
import { OldMaidPage } from './pages/OldMaidPage';
import { OmahaPage } from './pages/OmahaPage';
import { PaiGowPage } from './pages/PaiGowPage';
import { PigsTailPage } from './pages/PigsTailPage';
import { PineapplePage } from './pages/PineapplePage';
import { PinochlePage } from './pages/PinochlePage';
import { PokerPage } from './pages/PokerPage';
import { PokerSquaresPage } from './pages/PokerSquaresPage';
import { PyramidPage } from './pages/PyramidPage';
import { SevenCardStudPage } from './pages/SevenCardStudPage';
import { SevensPage } from './pages/SevensPage';
import { ShortDeckPage } from './pages/ShortDeckPage';
import { SpadesPage } from './pages/SpadesPage';
import { SpeedPage } from './pages/SpeedPage';
import { SpiderPage } from './pages/SpiderPage';
import { ThreeCardPage } from './pages/ThreeCardPage';
import { TriPeaksPage } from './pages/TriPeaksPage';
import { TwoTenJackPage } from './pages/TwoTenJackPage';
import { VideoPokerPage } from './pages/VideoPokerPage';
import { WarPage } from './pages/WarPage';
import { WhistPage } from './pages/WhistPage';
import { YukonPage } from './pages/YukonPage';

type GamePath = (typeof gameRoutes)[number]['path'];
const pageByPath: Record<GamePath, ReactNode> = {
  '/': <BlackJackPage />,
  '/poker': <PokerPage />,
  '/oldmaid': <OldMaidPage />,
  '/daifugo': <DaifugoPage />,
  '/sevens': <SevensPage />,
  '/doubt': <DoubtPage />,
  '/durak': <DurakPage />,
  '/euchre': <EuchrePage />,
  '/bridge': <BridgePage />,
  '/holdem': <HoldemPage />,
  '/omaha': <OmahaPage />,
  '/pineapple': <PineapplePage />,
  '/sevencardstud': <SevenCardStudPage />,
  '/shortdeck': <ShortDeckPage />,
  '/hearts': <HeartsPage />,
  '/spades': <SpadesPage />,
  '/twotenjack': <TwoTenJackPage />,
  '/ohhell': <OhHellPage />,
  '/napoleon': <NapoleonPage />,
  '/memory': <MemoryPage />,
  '/klondike': <KlondikePage />,
  '/freecell': <FreeCellPage />,
  '/baccarat': <BaccaratPage />,
  '/crazyeights': <CrazyEightsPage />,
  '/ginrummy': <GinRummyPage />,
  '/canasta': <CanastaPage />,
  '/cribbage': <CribbagePage />,
  '/spider': <SpiderPage />,
  '/pyramid': <PyramidPage />,
  '/tripeaks': <TriPeaksPage />,
  '/indianpoker': <IndianPokerPage />,
  '/videopoker': <VideoPokerPage />,
  '/deuceswild': <DeucesWildPage />,
  '/jokerpoker': <JokerPokerPage />,
  '/threecard': <ThreeCardPage />,
  '/caribbeanstud': <CaribbeanStudPage />,
  '/paigow': <PaiGowPage />,
  '/letitride': <LetItRidePage />,
  '/speed': <SpeedPage />,
  '/gofish': <GoFishPage />,
  '/pinochle': <PinochlePage />,
  '/golf': <GolfPage />,
  '/pigtail': <PigsTailPage />,
  '/war': <WarPage />,
  '/fiftyone': <FiftyOnePage />,
  '/clocksolitaire': <ClockSolitairePage />,
  '/fortythieves': <FortyThievesPage />,
  '/canfield': <CanfieldPage />,
  '/yukon': <YukonPage />,
  '/whist': <WhistPage />,
  '/poker-squares': <PokerSquaresPage />,
};

/** Root application component with router and game page routes. */
export default function App() {
  const { t } = useTranslation();
  return (
    <HashRouter>
      <ErrorBoundary>
        <div className="flex flex-col h-full lg:flex-row">
          <SkipNavLink targetId="main-content" label={t('nav.skipToContent')} />
          <DesktopSidebar />
          <div className="flex flex-col flex-1 min-w-0">
            <NavBar />
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
