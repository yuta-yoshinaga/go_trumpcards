import { HashRouter, Route, Routes } from 'react-router-dom';
import { NavBar } from './components/NavBar';
import { BlackJackPage } from './pages/BlackJackPage';
import { DaifugoPage } from './pages/DaifugoPage';
import { OldMaidPage } from './pages/OldMaidPage';
import { PokerPage } from './pages/PokerPage';
import { SevensPage } from './pages/SevensPage';

export default function App() {
  return (
    <HashRouter>
      <div className="flex flex-col h-full">
        <NavBar />
        <div className="flex-1 flex flex-col min-h-0">
          <Routes>
            <Route path="/" element={<BlackJackPage />} />
            <Route path="/poker" element={<PokerPage />} />
            <Route path="/oldmaid" element={<OldMaidPage />} />
            <Route path="/daifugo" element={<DaifugoPage />} />
            <Route path="/sevens" element={<SevensPage />} />
          </Routes>
        </div>
      </div>
    </HashRouter>
  );
}
