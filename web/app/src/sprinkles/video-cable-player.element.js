import {
  getActionCableConsumer,
  attachCSRFToken,
  makeWebSocketURL,
} from '@sargassum-world/stimulated';

export default class VideoCablePlayerElement extends HTMLCanvasElement {
  async connectedCallback() {
    if (document.documentElement.hasAttribute('data-turbo-preview')) {
      return;
    }
    await attachCSRFToken(this);

    // Initialize channel
    const channel = {
      channel: this.getAttribute('channel'),
      name: this.getAttribute('name'),
      integrity: this.getAttribute('integrity'),
    };
    if (this.hasAttribute('csrf-token')) {
      channel.csrfToken = this.getAttribute('csrf-token');
    }

    // Subscribe
    const consumer = getActionCableConsumer(
      makeWebSocketURL(this.getAttribute('cable-route')),
      // Warning: VideoCablePlayerElement assumes the use of a format which can serialize byte
      // arrays, such as MessagePack - but not JSON!
      this.getAttribute('websocket-subprotocol'),
    );
    this.subscription = consumer.subscriptions.create(channel, {
      received: this.dispatchMessageEvent.bind(this),
    });
  }

  disconnectedCallback() {
    if (this.subscription) this.subscription.unsubscribe();
  }

  async dispatchMessageEvent(data) {
    const bitmap = await createImageBitmap(
      new Blob([data], { type: 'image/jpeg' }),
    );
    this.width = bitmap.width;
    this.height = bitmap.height;
    this.getContext('2d').drawImage(bitmap, 0, 0);
  }
}
