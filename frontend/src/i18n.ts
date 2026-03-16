import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";
import translation_en from "@/locales/en-US/translation.json";
import translation_zh from "@/locales/zh-CN/translation.json";

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    debug: false,
    supportedLngs: ["zh-CN", "en-US"],
    fallbackLng: "zh-CN",
    load: "currentOnly",
    detection: {
      order: ["localStorage"],
      caches: ["localStorage"],
    },
    interpolation: {
      escapeValue: false,
    },
    resources: {
      "en-US": {
        translation: translation_en,
      },
      "zh-CN": {
        translation: translation_zh,
      },
    },
  });

export default i18n;

export const lngs = [
  {
    code: "zh-CN",
    nativeName: "简体中文",
    switchText: "中",
    officialSiteCode: "zh-cn",
  },
  {
    code: "en-US",
    nativeName: "English (United States)",
    switchText: "En",
    officialSiteCode: "en-us",
  },
];
