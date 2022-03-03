import { Controller } from 'stimulus';

export default class extends Controller {
  static targets = ['submit', 'submitter'];

  submit() {
    this.submitTarget.setAttribute('disabled', true);
    const progress = document.createElement('progress');
    progress.classList.add('progress');
    progress.classList.add('is-small');
    progress.classList.add('is-info');
    this.submitterTarget.appendChild(progress);
  }
}
