type Theme = 'dark' | 'light';

export function loadThemeSetting(): Theme {
  const theme = window.localStorage.getItem('theme');
  switch (theme) {
    case 'dark':
      return 'dark';
    default:
      return 'light';
  }
}
export function storeThemeSetting(theme: Theme): void {
  window.localStorage.setItem('theme', theme);
}
export function invert(theme: Theme): Theme {
  switch (theme) {
    case 'dark':
      return 'light';
    default:
      return 'dark';
  }
}
function setThemeProperties(stylesheetID: string, theme: Theme): void {
  const stylesheet = document.querySelector(`link#${stylesheetID}`);
  if (stylesheet === null || !(stylesheet instanceof HTMLLinkElement)) {
    return;
  }

  const active = stylesheet.id === `${theme}-theme`;
  stylesheet.rel = active ? 'stylesheet' : 'stylesheet alternate';
}
export function setTheme(theme: Theme): void {
  setThemeProperties('light-theme', theme);
  setThemeProperties('dark-theme', theme);
}

export function onLoad(): void {
  const theme = loadThemeSetting();
  if (theme === 'dark') {
    // Set an initial color to prevent a white flash when dark mode has been set
    document.documentElement.classList.add('initial-load-dark');
    const darkStylesheet = document.querySelector('link#dark-theme');
    if (darkStylesheet !== null && darkStylesheet instanceof HTMLLinkElement) {
      darkStylesheet.addEventListener('load', () => {
        document.documentElement.classList.remove('initial-load-dark');
      });
    }
  }
  setTheme(theme);
  storeThemeSetting(theme);
}
