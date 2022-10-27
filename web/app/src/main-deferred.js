import {
  CSRFController,
  DefaultScrollableController,
  FormSubmissionController,
  HideableController,
  LoadFocusController,
  LoadScrollController,
  NavigationLinkController,
  NavigationMenuController,
  ThemeController,
  TurboCableStreamSourceElement,
  TurboCacheController,
  ImageAutoreloadController,
  Turbo,
} from '@sargassum-world/stimulated';
import { Application } from '@hotwired/stimulus';

Turbo.session.drive = true;

customElements.define(
  'turbo-cable-stream-source',
  TurboCableStreamSourceElement,
);

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

export {};
