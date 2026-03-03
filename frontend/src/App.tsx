import type { ReactNode } from 'react';
import { HashRouter, Route, Routes } from 'react-router-dom';
import { NavBar } from './components/NavBar';
import { gameRoutes } from './constants/gameRoutes';
import { BlackJackPage } from './pages/BlackJackPage';
import { DaifugoPage } from './pages/DaifugoPage';
import { DoubtPage } from './pages/DoubtPage';
import { HoldemPage } from './pages/HoldemPage';
import { OldMaidPage } from './pages/OldMaidPage';
import { PokerPage } from './pages/PokerPage';
import { SevensPage } from './pages/SevensPage';

const pageByPath: Record<string, ReactNode> = {
  '/': <BlackJackPage />,
  '/poker': <PokerPage />,
  '/oldmaid': <OldMaidPage />,
  '/daifugo': <DaifugoPage />,
  '/sevens': <SevensPage />,
  '/doubt': <DoubtPage />,
  '/holdem': <HoldemPage />,
};

export default function App() {
  return (
    <HashRouter>
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
    </HashRouter>
  );
}
