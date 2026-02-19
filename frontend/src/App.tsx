import { HashRouter, Routes, Route } from 'react-router-dom'
import { NavBar } from './components/NavBar'
import { BlackJackPage } from './pages/BlackJackPage'
import { PokerPage } from './pages/PokerPage'
import { OldMaidPage } from './pages/OldMaidPage'
import { DaifugoPage } from './pages/DaifugoPage'

export default function App() {
  return (
    <HashRouter>
      <NavBar />
      <Routes>
        <Route path="/" element={<BlackJackPage />} />
        <Route path="/poker" element={<PokerPage />} />
        <Route path="/oldmaid" element={<OldMaidPage />} />
        <Route path="/daifugo" element={<DaifugoPage />} />
      </Routes>
    </HashRouter>
  )
}
