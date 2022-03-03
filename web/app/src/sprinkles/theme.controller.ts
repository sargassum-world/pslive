import { Controller } from 'stimulus';

import {
  invert,
  loadThemeSetting,
  storeThemeSetting,
  setTheme,
} from './theme-toggle';

export default class extends Controller {
  connect(): void {
    this.element.classList.remove('is-hidden');
  }
  toggle(): void {
    const theme = invert(loadThemeSetting());
    setTheme(theme);
    storeThemeSetting(theme);
  }
}
