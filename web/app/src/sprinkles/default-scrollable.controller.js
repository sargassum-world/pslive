import { Controller } from 'stimulus';

export default class extends Controller {
  connect() {
    this.element.focus();
    var scrollpos = sessionStorage.getItem('scrollpos');
    // TODO: instead of setting scrollTop (which causes a flash of unscrolled content,
    // change the HTML element structure of the navbar so that the main container just
    // has the window scroll, to let the browser natively keep track of the scroll
    // position when the page is refreshed
    if (scrollpos) {
      this.element.scrollTop = scrollpos;
      sessionStorage.removeItem('scrollpos');
    }

    const element = this.element;
    window.addEventListener('beforeunload', function () {
      sessionStorage.setItem('scrollpos', element.scrollTop);
    })
  }
}
