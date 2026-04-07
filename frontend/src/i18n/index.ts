import i18n from 'i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import { initReactI18next } from 'react-i18next';

/** Eagerly import all locale JSON files via Vite glob. */
const jaModules = import.meta.glob('./locales/ja/*.json', {
  eager: true,
  import: 'default',
}) as Record<string, Record<string, string>>;

const enModules = import.meta.glob('./locales/en/*.json', {
  eager: true,
  import: 'default',
}) as Record<string, Record<string, string>>;

/** Extract namespace name from glob path (e.g. './locales/ja/blackjack.json' → 'blackjack'). */
function buildResources(modules: Record<string, Record<string, string>>): Record<string, Record<string, string>> {
  const resources: Record<string, Record<string, string>> = {};
  for (const [path, mod] of Object.entries(modules)) {
    const name = path.split('/').pop()?.replace('.json', '');
    if (name) resources[name] = mod;
  }
  return resources;
}

const jaResources = buildResources(jaModules);
const enResources = buildResources(enModules);

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      ja: jaResources,
      en: enResources,
    },
    fallbackLng: 'ja',
    defaultNS: 'common',
    ns: Object.keys(jaResources),
    detection: {
      order: ['localStorage'],
      lookupLocalStorage: 'i18n_lang',
      caches: ['localStorage'],
    },
    interpolation: {
      escapeValue: false,
    },
  });

// Sync <html lang> with the active i18n language (WCAG 2.1 SC 3.1.1)
if (typeof document !== 'undefined') {
  document.documentElement.lang = i18n.language;
  i18n.on('languageChanged', (lng: string) => {
    document.documentElement.lang = lng;
  });
}

/** Configured i18next instance with ja/en translations and language detection. */
export default i18n;
