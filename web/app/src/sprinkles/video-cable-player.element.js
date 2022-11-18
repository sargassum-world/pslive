// This is adapted from the MIT-licensed library @hotwired/turbo-rails.
import { createConsumer, logger } from '@rails/actioncable';
import {
  getCSRFToken,
  setCSRFToken,
  fetchCSRFToken,
} from '@sargassum-world/stimulated';

let consumers = {};

function getConsumer(url) {
  if (consumers[url] === undefined) {
    consumers[url] = createConsumer(url === null ? undefined : url);
  }
  return consumers[url];
}

export default class VideoCablePlayerElement extends HTMLCanvasElement {
  async connectedCallback() {
    if (this.hasValidCSRFToken()) {
      setCSRFToken(this.getAttribute('csrf-token'));
    } else if (
      this.hasAttribute('csrf-token') &&
      this.hasAttribute('data-csrf-route')
    ) {
      await this.addCSRFToken();
    }
    const channel = {
      channel: this.getAttribute('channel'),
      name: this.getAttribute('name'),
      integrity: this.getAttribute('integrity'),
    };
    if (this.hasAttribute('csrf-token')) {
      channel.csrfToken = this.getAttribute('csrf-token');
    }
    this.subscription = getConsumer(
      this.getAttribute('cable-route'),
    ).subscriptions.create(channel, {
      received: this.dispatchMessageEvent.bind(this),
    });
    if (this.hasAttribute('logging')) {
      logger.enabled = true;
    }
  }

  disconnectedCallback() {
    if (this.subscription) this.subscription.unsubscribe();
  }

  async dispatchMessageEvent(data) {
    const decoded = atob(data);
    const array = new Uint8Array(decoded.length);
    for (var i = 0; i < decoded.length; i++) {
      array[i] = decoded.charCodeAt(i);
    }
    const bitmap = await createImageBitmap(
      new Blob([array], { type: 'image/jpeg' }),
    );
    this.width = bitmap.width;
    this.height = bitmap.height;
    this.getContext('2d').drawImage(bitmap, 0, 0);
  }

  async addCSRFToken() {
    if (this.hasValidCSRFToken() || this.setCSRFToken()) {
      return;
    }
    await fetchCSRFToken(this.getAttribute('data-csrf-route'));
    this.setCSRFToken();
  }

  setCSRFToken() {
    if (getCSRFToken() === null) {
      return false;
    }
    this.setAttribute('csrf-token', getCSRFToken());
    return this.hasValidCSRFToken();
  }

  hasValidCSRFToken() {
    return this.getAttribute('csrf-token').length > 0;
  }
}
