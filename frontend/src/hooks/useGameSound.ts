import { useCallback, useEffect, useState } from 'react';

const SOUND_MUTED_KEY = 'trumpcards-sound-muted';

/** Hook providing game sound effects with global mute toggle persisted to localStorage. */
export function useGameSound() {
  const [muted, setMuted] = useState(() => {
    try {
      return localStorage.getItem(SOUND_MUTED_KEY) === 'true';
    } catch {
      return false;
    }
  });

  useEffect(() => {
    try {
      localStorage.setItem(SOUND_MUTED_KEY, String(muted));
    } catch {
      // localStorage unavailable
    }
  }, [muted]);

  const toggleMute = useCallback(() => setMuted((prev) => !prev), []);

  const playSound = useCallback(
    (src: string) => {
      if (muted) return;
      try {
        const audio = new Audio(src);
        audio.volume = 0.3;
        audio.play().catch(() => {
          // Autoplay blocked — ignore
        });
      } catch {
        // Audio unavailable
      }
    },
    [muted],
  );

  const playCardDeal = useCallback(() => playSound('/sounds/card-deal.ogg'), [playSound]);
  const playCardFlip = useCallback(() => playSound('/sounds/card-flip.ogg'), [playSound]);
  const playSelect = useCallback(() => playSound('/sounds/card-select.ogg'), [playSound]);
  const playWin = useCallback(() => playSound('/sounds/win.ogg'), [playSound]);

  return { muted, toggleMute, playCardDeal, playCardFlip, playSelect, playWin };
}
