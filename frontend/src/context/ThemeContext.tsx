import { createContext, useContext, useEffect, useState, ReactNode } from "react";

type ThemeName = "dark" | "light" | "default";

type Palette = {
  background: string;
  foreground: string;
  card: string;
  'card-foreground': string;
  primary: string;
  'primary-foreground': string;
  secondary: string;
  'secondary-foreground': string;
  accent: string;
  'accent-foreground': string;
  destructive: string;
  'destructive-foreground': string;
  border: string;
  input: string;
  muted: string;
  'muted-foreground': string;
  popover: string;
  'popover-foreground': string;
  ring: string;
};

interface ThemeContextType {
  theme: ThemeName;
  setTheme: (t: ThemeName) => void;
  palette: Palette;
  setPalette: (p: Partial<Palette>) => void;
}

const basePalettes: Record<ThemeName, Palette> = {
  dark: {
    background: "220 40% 15%", // Match card color
    foreground: "213 31% 91%",
    card: "220 40% 15%",
    'card-foreground': "213 31% 91%",
    primary: "213 96% 53%",
    'primary-foreground': "210 40% 98%",
    secondary: "220 40% 35%",
    'secondary-foreground': "213 31% 91%",
    accent: "216 34% 17%",
    'accent-foreground': "210 40% 98%",
    destructive: "0 63% 31%",
    'destructive-foreground': "210 40% 98%",
    border: "216 34% 17%",
    input: "216 34% 17%",
    muted: "223 47% 11%",
    'muted-foreground': "215.4 16.3% 56.9%",
    popover: "224 71% 4%",
    'popover-foreground': "215 20.2% 65.1%",
    ring: "216 34% 17%",
  },
  light: {
    background: "45 56% 96%", // IVORY
    foreground: "120 13% 30%", // SAGE (darker for contrast)
    card: "24 33% 94%", // NUDE
    'card-foreground': "120 13% 30%",
    primary: "205 79% 70%", // BABY BLUE
    'primary-foreground': "120 13% 30%",
    secondary: "340 26% 76%", // DUSTY ROSE
    'secondary-foreground': "120 13% 30%",
    accent: "120 13% 86%", // SAGE
    'accent-foreground': "120 13% 30%",
    destructive: "0 84% 60%",
    'destructive-foreground': "120 13% 30%",
    border: "24 33% 80%",
    input: "24 33% 80%",
    muted: "45 56% 90%",
    'muted-foreground': "120 13% 50%",
    popover: "45 56% 96%",
    'popover-foreground': "120 13% 30%",
    ring: "205 79% 70%",
  },
  default: {
    background: "220 14% 96%",
    foreground: "222 47% 11%",
    card: "210 20% 98%",
    'card-foreground': "222 47% 11%",
    primary: "213 96% 53%",
    'primary-foreground': "210 40% 98%",
    secondary: "215 19% 35%",
    'secondary-foreground': "210 40% 98%",
    accent: "210 40% 96.1%",
    'accent-foreground': "222.2 47.4% 11.2%",
    destructive: "0 100% 50%",
    'destructive-foreground': "210 40% 98%",
    border: "214.3 31.8% 91.4%",
    input: "214.3 31.8% 91.4%",
    muted: "210 40% 96.1%",
    'muted-foreground': "215.4 16.3% 46.9%",
    popover: "0 0% 100%",
    'popover-foreground': "222.2 47.4% 11.2%",
    ring: "215 20.2% 65.1%",
  },
};

const getPaletteOverride = (theme: ThemeName): Partial<Palette> => {
  if (typeof window !== 'undefined') {
    const val = localStorage.getItem(`palette-override-${theme}`);
    if (val) {
      try {
        return JSON.parse(val);
      } catch {}
    }
  }
  return {};
};

const getInitialTheme = (): ThemeName => {
  if (typeof window !== 'undefined') {
    const t = localStorage.getItem('theme') as ThemeName | null;
    if (t === 'dark' || t === 'light' || t === 'default') return t;
  }
  return 'default';
};

const getInitialPalette = (theme: ThemeName): Palette => {
  return {
    ...basePalettes[theme],
    ...getPaletteOverride(theme),
  };
};

const ThemeContext = createContext<ThemeContextType>({
  theme: 'default',
  setTheme: () => {},
  palette: basePalettes.default,
  setPalette: () => {},
});

export const ThemeProvider = ({ children }: { children: ReactNode }) => {
  // On mount, get theme from class or localStorage
  const getInitialThemeFromClass = (): ThemeName => {
    if (typeof window !== 'undefined') {
      const html = document.documentElement;
      if (html.classList.contains('dark')) return 'dark';
      if (html.classList.contains('light')) return 'light';
      if (html.classList.contains('default')) return 'default';
    }
    return getInitialTheme();
  };

  const [theme, setThemeState] = useState<ThemeName>(getInitialThemeFromClass());
  const [palette, setPaletteState] = useState<Palette>(getInitialPalette(getInitialThemeFromClass()));

  // Set theme class and palette variables together
  const applyTheme = (theme: ThemeName, palette: Palette) => {
    const root = document.documentElement;
    root.classList.remove('dark', 'light', 'default');
    root.classList.add(theme);
    Object.entries(palette).forEach(([key, value]) => {
      root.style.setProperty(`--${key}`, value);
    });
  };

  // On mount and whenever theme or palette changes
  useEffect(() => {
    applyTheme(theme, palette);
  }, [palette, theme]);

  // When theme changes, update palette and localStorage
  useEffect(() => {
    const newPalette = {
      ...basePalettes[theme],
      ...getPaletteOverride(theme),
    };
    setPaletteState(newPalette);
    if (typeof window !== 'undefined') {
      localStorage.setItem('theme', theme);
    }
  }, [theme]);

  const setPalette = (p: Partial<Palette>) => {
    setPaletteState(prev => {
      const next = { ...prev, ...p };
      applyTheme(theme, next);
      if (typeof window !== 'undefined') {
        localStorage.setItem(`palette-override-${theme}`, JSON.stringify(next));
      }
      return next;
    });
  };

  const setTheme = (t: ThemeName) => {
    setThemeState(t);
    // Immediately update <html> class and CSS variables for instant UI feedback
    const newPalette = {
      ...basePalettes[t],
      ...getPaletteOverride(t),
    };
    const root = document.documentElement;
    root.classList.remove('dark', 'light', 'default');
    root.classList.add(t);
    Object.entries(newPalette).forEach(([key, value]) => {
      root.style.setProperty(`--${key}`, value);
    });
    if (typeof window !== 'undefined') {
      localStorage.setItem('theme', t);
      localStorage.setItem(`palette-override-${t}`, JSON.stringify(newPalette));
    }
  };

  return (
    <ThemeContext.Provider value={{ theme, setTheme, palette, setPalette }}>
      {children}
    </ThemeContext.Provider>
  );
};

export const useTheme = () => useContext(ThemeContext);
