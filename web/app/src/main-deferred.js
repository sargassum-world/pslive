import * as Turbo from '@hotwired/turbo';
import {
	CSRFController,
	DefaultScrollableController,
	FormSubmissionController,
	NavigationLinkController,
	NavigationMenuController,
	ThemeController,
	TurboCacheController,
} from '@sargassum-world/stimulated';
import { Application } from 'stimulus';

Turbo.session.drive = true;

const Stimulus = Application.start();
Stimulus.register('csrf', CSRFController);
Stimulus.register('default-scrollable', DefaultScrollableController);
Stimulus.register('form-submission', FormSubmissionController);
Stimulus.register('navigation-link', NavigationLinkController);
Stimulus.register('navigation-menu', NavigationMenuController);
Stimulus.register('theme', ThemeController);
Stimulus.register('turbo-cache', TurboCacheController);

export {};
