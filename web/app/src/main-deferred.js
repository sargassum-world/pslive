import {
  CSRFController,
  DefaultScrollableController,
  FormSubmissionController,
  HideableController,
  ImageAutoreloadController,
  LoadFocusController,
  LoadScrollController,
  NavigationLinkController,
  NavigationMenuController,
  ThemeController,
  TurboCableStreamSourceElement,
  TurboCacheController,
  Turbo,
} from '@sargassum-world/stimulated';
import { Application } from '@hotwired/stimulus';
import { VideoCablePlayerElement } from './sprinkles';

Turbo.session.drive = true;

customElements.define(
  'turbo-cable-stream-source',
  TurboCableStreamSourceElement,
);
customElements.define('video-cable-player', VideoCablePlayerElement, {
  extends: 'canvas',
});

const Stimulus = Application.start();
Stimulus.register('csrf', CSRFController);
Stimulus.register('default-scrollable', DefaultScrollableController);
Stimulus.register('form-submission', FormSubmissionController);
Stimulus.register('hideable', HideableController);
Stimulus.register('image-autoreload', ImageAutoreloadController);
Stimulus.register('load-focus', LoadFocusController);
Stimulus.register('load-scroll', LoadScrollController);
Stimulus.register('navigation-link', NavigationLinkController);
Stimulus.register('navigation-menu', NavigationMenuController);
Stimulus.register('theme', ThemeController);
Stimulus.register('turbo-cache', TurboCacheController);

if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/sw.js');
}

// Prevent noscript elements from being processed. Refer to
// https://discuss.hotwired.dev/t/turbo-processes-noscript-children-when-merging-head/2552
document.addEventListener('turbo:before-render', function (event) {
  event.detail.newBody.querySelectorAll('noscript').forEach((e) => e.remove());
});

export {};
