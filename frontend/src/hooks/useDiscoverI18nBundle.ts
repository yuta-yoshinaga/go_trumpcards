import i18n from 'i18next';
import { useEffect, useState } from 'react';

const NAMESPACE = 'discover';

/** Module cache so the dynamic imports only run once per language per session. */
const inFlight = new Map<string, Promise<void>>();

/**
 * Dynamically import the `discover` i18n bundle for the current language
 * and register it via `i18n.addResourceBundle`. Returns `true` once the
 * resource bundle is ready (or was already present), so the caller can
 * render a skeleton while it loads.
 *
 * The discover bundle is excluded from the eager glob in `buildResources`
 * so users who never visit `/discover` do not pay its ~25–35 KB gzipped
 * weight. Loading is keyed by language; switching languages mid-session
 * triggers a fresh import for the new locale.
 */
export function useDiscoverI18nBundle(): boolean {
  const [, force] = useState(0);
  const lang = i18n.language || 'ja';
  const ready = i18n.hasResourceBundle(lang, NAMESPACE);

  useEffect(() => {
    if (i18n.hasResourceBundle(lang, NAMESPACE)) return;
    const key = lang;
    const cached = inFlight.get(key);
    if (cached) {
      cached.then(() => force((n) => n + 1));
      return;
    }
    const load = import(`../i18n/locales/${lang}/discover.json`)
      .then((mod: { default: Record<string, unknown> }) => {
        i18n.addResourceBundle(lang, NAMESPACE, mod.default, true, true);
      })
      .catch((err: unknown) => {
        if (import.meta.env.DEV) {
          console.warn(`[useDiscoverI18nBundle] failed to load ${lang}:`, err);
        }
      })
      .finally(() => {
        inFlight.delete(key);
        force((n) => n + 1);
      });
    inFlight.set(key, load);
  }, [lang]);

  return ready;
}
