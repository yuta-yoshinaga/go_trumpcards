import { Howl } from 'howler';
import type { ReactNode } from 'react';
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';

const SOUND_MUTED_KEY = 'trumpcards-sound-muted';

/** Sound name identifier for all game sounds. */
export type SoundName =
  | 'cardDeal'
  | 'cardFlip'
  | 'cardSelect'
  | 'cardPlace'
  | 'shuffle'
  | 'chipClick'
  | 'winFanfare'
  | 'lossThud'
  | 'errorBuzz'
  | 'turnTick';

/** Options for playSound calls. */
export interface PlayOptions {
  /** Pitch randomization range. E.g., 0.05 = playbackRate 0.95-1.05. */
  pitchVariation?: number;
  /** Volume override, 0.0-1.0. Overrides the default tier volume. */
  volume?: number;
}

/** Context value provided by SoundProvider. */
interface SoundContextValue {
  /** Fire-and-forget sound playback. Silently no-ops if muted or sound fails. */
  playSound: (name: SoundName, options?: PlayOptions) => void;
  /** Whether sound is currently muted. */
  muted: boolean;
  /** Toggle mute state. Persists to localStorage. */
  toggleMute: () => void;
}

const SoundContext = createContext<SoundContextValue | null>(null);

/** Base volume for all sounds. */
const BASE_VOLUME = 0.3;

/** Volume multiplier per tier: ambient (background), action (user feedback), event (state change). */
const VOLUME_TIERS: Record<SoundName, number> = {
  cardDeal: 0.6,
  cardPlace: 0.6,
  turnTick: 0.6,
  cardSelect: 0.6,
  cardFlip: 1.0,
  chipClick: 1.0,
  shuffle: 1.0,
  winFanfare: 1.4,
  lossThud: 1.4,
  errorBuzz: 1.4,
};

/** Sound file paths mapped by name. */
const SOUND_FILES: Record<SoundName, string[]> = {
  cardDeal: ['/sounds/card-deal.ogg', '/sounds/card-deal.mp3'],
  cardFlip: ['/sounds/card-flip.ogg', '/sounds/card-flip.mp3'],
  cardSelect: ['/sounds/card-select.ogg', '/sounds/card-select.mp3'],
  cardPlace: ['/sounds/card-place.ogg', '/sounds/card-place.mp3'],
  shuffle: ['/sounds/shuffle.ogg', '/sounds/shuffle.mp3'],
  chipClick: ['/sounds/chip-click.ogg', '/sounds/chip-click.mp3'],
  winFanfare: ['/sounds/win-fanfare.ogg', '/sounds/win-fanfare.mp3'],
  lossThud: ['/sounds/loss-thud.ogg', '/sounds/loss-thud.mp3'],
  errorBuzz: ['/sounds/error-buzz.ogg', '/sounds/error-buzz.mp3'],
  turnTick: ['/sounds/turn-tick.ogg', '/sounds/turn-tick.mp3'],
};

/** Creates all Howl instances for preloading. Returns null on failure. */
function createSounds(): Record<SoundName, Howl> | null {
  try {
    const sounds = {} as Record<SoundName, Howl>;
    for (const [name, src] of Object.entries(SOUND_FILES)) {
      sounds[name as SoundName] = new Howl({
        src,
        preload: true,
        volume: BASE_VOLUME * VOLUME_TIERS[name as SoundName],
      });
    }
    return sounds;
  } catch {
    return null;
  }
}

/** Provides game sound context to the application. */
export function SoundProvider({ children }: { children: ReactNode }) {
  const [muted, setMuted] = useState(() => {
    try {
      return localStorage.getItem(SOUND_MUTED_KEY) === 'true';
    } catch {
      return false;
    }
  });

  const sounds = useMemo(() => createSounds(), []);

  useEffect(() => {
    try {
      localStorage.setItem(SOUND_MUTED_KEY, String(muted));
    } catch {
      // localStorage unavailable
    }
  }, [muted]);

  const toggleMute = useCallback(() => setMuted((prev) => !prev), []);

  const playSound = useCallback(
    (name: SoundName, options?: PlayOptions) => {
      if (muted || !sounds) return;

      const howl = sounds[name];
      if (!howl) return;

      const id = howl.play();

      // Apply volume (tier default or override)
      const vol = options?.volume ?? BASE_VOLUME * VOLUME_TIERS[name];
      howl.volume(vol, id);

      // Apply pitch variation
      if (options?.pitchVariation) {
        const variation = options.pitchVariation;
        const rate = 1 + (Math.random() * 2 - 1) * variation;
        howl.rate(rate, id);
      }
    },
    [muted, sounds],
  );

  const value = useMemo(() => ({ playSound, muted, toggleMute }), [playSound, muted, toggleMute]);

  return <SoundContext.Provider value={value}>{children}</SoundContext.Provider>;
}

/** Hook to access sound playback from any component within SoundProvider. */
export function useSound(): SoundContextValue {
  const context = useContext(SoundContext);
  if (!context) {
    throw new Error('useSound must be used within a SoundProvider');
  }
  return context;
}
