import { Controller } from 'stimulus';

export default class extends Controller {
  static targets = ['toggle', 'close', 'menu'];
  connect() {
    this.closeTarget.classList.add('is-hidden');
    this.toggleTarget.setAttribute('href', '#');
  }

  open() {
    this.menuTarget.classList.add('is-active');
    this.toggleTarget.classList.add('is-active');
    this.toggleTarget.setAttribute('aria-expanded', true);
  }

  close() {
    this.menuTarget.classList.remove('is-active');
    this.toggleTarget.classList.remove('is-active');
    this.toggleTarget.setAttribute('aria-expanded', false);
  }

  toggle(event) {
    if (this.menuTarget.classList.contains('is-active')) {
      this.close();
    } else {
      this.open();
    }
    event.preventDefault();
  }
}
