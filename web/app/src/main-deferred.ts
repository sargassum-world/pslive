import * as Turbo from '@hotwired/turbo';
import { Application } from 'stimulus';

import ThemeController from './sprinkles/theme.controller';
import NavigationMenuController from './sprinkles/navigation-menu.controller';
import NavigationLinkController from './sprinkles/navigation-link.controller';
import FormSubmissionController from './sprinkles/form-submission.controller';
import CSRFController from './sprinkles/csrf.controller';
import DefaultScrollableController from './sprinkles/default-scrollable.controller';
import TurboCacheController from './sprinkles/turbo-cache.controller';

Turbo.session.drive = true;

const Stimulus = Application.start();
Stimulus.register('theme', ThemeController);
Stimulus.register('navigation-menu', NavigationMenuController);
Stimulus.register('navigation-link', NavigationLinkController);
Stimulus.register('form-submission', FormSubmissionController);
Stimulus.register('csrf', CSRFController);
Stimulus.register('default-scrollable', DefaultScrollableController);
Stimulus.register('turbo-cache', TurboCacheController);

export {};
