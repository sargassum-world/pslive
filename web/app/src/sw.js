import {
  offlineFallback,
  staticResourceCache,
  warmStrategyCache,
} from 'workbox-recipes';
import { registerRoute, setDefaultHandler } from 'workbox-routing';
import { CacheFirst, NetworkOnly } from 'workbox-strategies';
import { ExpirationPlugin } from 'workbox-expiration';

const cacheFirst = new CacheFirst();

// Static assets (served with file-rev, so cache forever)

staticResourceCache();
registerRoute(/static\/.*/, cacheFirst);
registerRoute(/app\/.*/, cacheFirst);

// Favicon (served with max-age cache-control header)

const faviconCacheFirst = new CacheFirst({
  cacheName: 'favicon-cache',
  plugins: [
    new ExpirationPlugin({
      maxAgeSeconds: 1 * 24 * 60 * 60,
    }),
  ],
});
warmStrategyCache({
  urls: ['/favicon.ico'],
  strategy: faviconCacheFirst,
});
registerRoute(/favicon\.ico/, faviconCacheFirst);

// Fonts (served with max-age cache-control header)

const fontCacheFirst = new CacheFirst({
  cacheName: 'font-cache',
  plugins: [
    new ExpirationPlugin({
      maxAgeSeconds: 90 * 24 * 60 * 60,
    }),
  ],
});
warmStrategyCache({
  urls: [
    '/fonts/atkinson-hyperlegible-latin-400-normal.woff2',
    '/fonts/atkinson-hyperlegible-latin-700-normal.woff2',
  ],
  strategy: fontCacheFirst,
});
registerRoute(/fonts\/.*/, fontCacheFirst);

// Pages (no cache)

setDefaultHandler(new NetworkOnly());
warmStrategyCache({
  urls: ['/app/offline'],
  strategy: cacheFirst,
});
offlineFallback({
  pageFallback: '/app/offline',
});
