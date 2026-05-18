import i18n from 'i18next';
import { useEffect, useState } from 'react';

const NAMESPACE = 'discover';
/** Locales for which a `discover.json` bundle exists on disk. */
const SUPPORTED_LANGS = ['ja', 'en'] as const;
type SupportedLang = (typeof SUPPORTED_LANGS)[number];
const FALLBACK_LANG: SupportedLang = 'ja';

/** Module cache so the dynamic imports only run once per language per session. */
const inFlight = new Map<SupportedLang, Promise<void>>();

/**
 * Resolve an i18next language tag (which may be BCP-47, e.g. `en-US`)
 * down to one of the locales we actually ship — falling back to ja so
 * users on unsupported locales still get a real bundle instead of a
 * silent 404 + permanent skeleton.
 */
function resolveLang(raw: string | undefined): SupportedLang {
  const base = (raw || FALLBACK_LANG).split('-')[0];
  return (SUPPORTED_LANGS as readonly string[]).includes(base) ? (base as SupportedLang) : FALLBACK_LANG;
}

/** Vite-statically-analyzable glob of every shippable discover bundle. */
const discoverBundles = import.meta.glob<{ default: Record<string, unknown> }>('../i18n/locales/*/discover.json');

/** Resolve the loader function for a given supported lang. */
function loaderFor(lang: SupportedLang): () => Promise<{ default: Record<string, unknown> }> {
  const key = `../i18n/locales/${lang}/discover.json`;
  const loader = discoverBundles[key];
  if (!loader) throw new Error(`[useDiscoverI18nBundle] missing bundle for ${lang}`);
  return loader;
}

/**
 * Dynamically import the `discover` i18n bundle for the current language
 * and register it via `i18n.addResourceBundle`. Returns `true` once the
 * resource bundle is ready (or was already present), so the caller can
 * render a skeleton while it loads.
 *
 * The discover bundle is excluded from the eager glob in `buildResources`
 * so users who never visit `/discover` do not pay its ~25–35 KB gzipped
 * weight. Loading is keyed by the resolved supported language; on import
 * failure we fall back to the ja bundle so the user is never stranded on
 * the skeleton (network failures, missing locale, etc.).
 */
export function useDiscoverI18nBundle(): boolean {
  const [, force] = useState(0);
  const lang = resolveLang(i18n.language);
  const ready = i18n.hasResourceBundle(lang, NAMESPACE) || i18n.hasResourceBundle(FALLBACK_LANG, NAMESPACE);

  useEffect(() => {
    if (i18n.hasResourceBundle(lang, NAMESPACE)) return;
    const cached = inFlight.get(lang);
    if (cached) {
      cached.then(() => force((n) => n + 1));
      return;
    }
    const tryLoad = (target: SupportedLang) =>
      loaderFor(target)().then((mod) => {
        i18n.addResourceBundle(target, NAMESPACE, mod.default, true, true);
      });
    const load = tryLoad(lang)
      .catch((primaryErr: unknown) => {
        if (import.meta.env.DEV) {
          console.warn(`[useDiscoverI18nBundle] failed to load ${lang}, trying ${FALLBACK_LANG}:`, primaryErr);
        }
        if (lang === FALLBACK_LANG) return;
        return tryLoad(FALLBACK_LANG).catch((fallbackErr: unknown) => {
          if (import.meta.env.DEV) {
            console.warn(`[useDiscoverI18nBundle] fallback ${FALLBACK_LANG} also failed:`, fallbackErr);
          }
        });
      })
      .finally(() => {
        inFlight.delete(lang);
        force((n) => n + 1);
      });
    inFlight.set(lang, load);
  }, [lang]);

  return ready;
}

/**
 * Test-only: clear the module-level in-flight cache. Production code
 * never calls this; tests use it to isolate hook invocations.
 */
export function __resetDiscoverI18nBundleCacheForTests(): void {
  inFlight.clear();
}
