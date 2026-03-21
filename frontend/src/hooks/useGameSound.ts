import { useCallback, useEffect, useMemo, useState } from 'react';

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

  const sounds = useMemo(() => {
    try {
      const audioMap = {
        cardDeal: new Audio('/sounds/card-deal.ogg'),
        cardFlip: new Audio('/sounds/card-flip.ogg'),
        select: new Audio('/sounds/card-select.ogg'),
        win: new Audio('/sounds/win.ogg'),
      };
      for (const audio of Object.values(audioMap)) {
        audio.volume = 0.3;
      }
      return audioMap;
    } catch {
      return null;
    }
  }, []);

  const playSound = useCallback(
    (audio: HTMLAudioElement | undefined) => {
      if (muted || !audio) return;
      audio.currentTime = 0;
      audio.play().catch(() => {
        // Autoplay blocked — ignore
      });
    },
    [muted],
  );

  const playCardDeal = useCallback(() => playSound(sounds?.cardDeal), [playSound, sounds]);
  const playCardFlip = useCallback(() => playSound(sounds?.cardFlip), [playSound, sounds]);
  const playSelect = useCallback(() => playSound(sounds?.select), [playSound, sounds]);
  const playWin = useCallback(() => playSound(sounds?.win), [playSound, sounds]);

  return { muted, toggleMute, playCardDeal, playCardFlip, playSelect, playWin };
}
