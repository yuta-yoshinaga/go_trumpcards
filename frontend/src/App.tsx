import type { ReactNode } from 'react';
import { HashRouter, Route, Routes } from 'react-router-dom';
import { ErrorBoundary } from './components/ErrorBoundary';
import { NavBar } from './components/NavBar';
import { gameRoutes } from './constants/gameRoutes';
import { BaccaratPage } from './pages/BaccaratPage';
import { BlackJackPage } from './pages/BlackJackPage';
import { DaifugoPage } from './pages/DaifugoPage';
import { DoubtPage } from './pages/DoubtPage';
import { HeartsPage } from './pages/HeartsPage';
import { HoldemPage } from './pages/HoldemPage';
import { KlondikePage } from './pages/KlondikePage';
import { MemoryPage } from './pages/MemoryPage';
import { OldMaidPage } from './pages/OldMaidPage';
import { PokerPage } from './pages/PokerPage';
import { SevensPage } from './pages/SevensPage';

type GamePath = (typeof gameRoutes)[number]['path'];
const pageByPath: Record<GamePath, ReactNode> = {
  '/': <BlackJackPage />,
  '/poker': <PokerPage />,
  '/oldmaid': <OldMaidPage />,
  '/daifugo': <DaifugoPage />,
  '/sevens': <SevensPage />,
  '/doubt': <DoubtPage />,
  '/holdem': <HoldemPage />,
  '/hearts': <HeartsPage />,
  '/memory': <MemoryPage />,
  '/klondike': <KlondikePage />,
  '/baccarat': <BaccaratPage />,
};

export default function App() {
  return (
    <HashRouter>
      <ErrorBoundary>
        <div className="flex flex-col h-full">
          <NavBar />
          <div className="flex-1 flex flex-col min-h-0">
            <Routes>
              {gameRoutes.map(({ path }) => (
                <Route key={path} path={path} element={pageByPath[path]} />
              ))}
            </Routes>
          </div>
        </div>
      </ErrorBoundary>
    </HashRouter>
  );
}
