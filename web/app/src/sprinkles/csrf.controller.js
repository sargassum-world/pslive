import { Controller } from 'stimulus';
import { Mutex } from 'async-mutex';

// These are globals because we only need to fetch a CSRF token once; we can reuse it as long as
// the CSRF cookie (for the Double Submit Cookie pattern) remains valid.
let csrfToken = null;
let fetchMutex = new Mutex();

export default class extends Controller {
  static targets = ['token', 'route', 'omit', 'submit'];
  async connect() {
    this.omitTarget.value = true;
    if (this.hasValidToken()) {
      csrfToken = this.tokenTarget.value;
      return;
    }
    await this.addToken()
  }

  async addToken(e) {
    if (this.hasValidToken() || this.setToken()) {
      return;
    }

    if (e !== undefined) {
      e.preventDefault();
    }
    if (fetchMutex.isLocked()) {
      await fetchMutex.waitForUnlock();
    }
    if (this.setToken()) {
      if (e !== undefined) {
        e.target.submit();
      }
      return;
    }

    await fetchMutex.runExclusive(async () => {
      const response = await fetch(this.routeTarget.value);
      const responseBody = await response.json();
      csrfToken = responseBody.token;
      this.setToken();
    });
    if (e !== undefined) {
      // Assumes the controller is mounted on a form!
      e.target.submit();
    }
  }

  setToken() {
    if (csrfToken === null) {
      return false
    }
    this.tokenTarget.value = csrfToken;
    return this.hasValidToken()
  }

  hasValidToken() {
    return this.tokenTarget.value.length > 0
  }
}
