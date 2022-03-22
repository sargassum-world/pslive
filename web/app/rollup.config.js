import svelte from 'rollup-plugin-svelte';
import commonjs from '@rollup/plugin-commonjs';
import resolve from '@rollup/plugin-node-resolve';
import { terser } from 'rollup-plugin-terser';
import sveltePreprocess from 'svelte-preprocess';
import css from 'rollup-plugin-css-only';
import scss from 'rollup-plugin-scss';
import { existsSync, mkdirSync, writeFileSync } from 'fs';
import purify from 'purify-css';
import copy from 'rollup-plugin-copy';

const production = !process.env.ROLLUP_WATCH;
const buildDir = 'public/build';

function themeGenerator(theme) {
	return {
		input: `src/theme-${theme}.js`,
		output: {
			sourcemap: !production,
			format: 'iife',
			name: `appTheme_${theme}`,
			file: `${buildDir}/theme-${theme}.js`
		},
		plugins: [
			scss({
				includePaths: [
					'node_modules',
					'src'
				],
				runtime: require('sass'),
				output: function (styles, styleNodes) {
					if (!existsSync(buildDir)) {
						mkdirSync(buildDir, { recursive: true });
					}
					writeFileSync(`${buildDir}/theme-${theme}.css`, styles);
					purify(
						[
							'node_modules/@hotwired/**/*.js',
							'node_modules/@sargassum-world/**/*.js',
							'node_modules/@sargassum-world/**/*.svelte',
							'src/**/*.js',
							'src/**/*.svelte',
							'../templates/**/*.tmpl',
						],
						styles,
						{
							output: `${buildDir}/theme-${theme}.min.css`,
							minify: true,
							info: true,
							cleanCssOptions: {
								level: {
									2: {
										all: true,
									},
								},
							},
						},
					);
				},
			}),
		],
		watch: {
			clearScreen: false
		}
	};
}

function bundleGenerator(type, appName, context) {
	return {
		input: `src/main-${type}.js`,
		output: {
			sourcemap: !production,
			format: 'iife',
			name: appName,
			file: `${buildDir}/bundle-${type}.js`
		},
		context,
		plugins: [
			svelte({
				preprocess: sveltePreprocess({
					sourceMap: !production,
					transformers: {
						scss: {
							includePaths: [
								'node_modules',
								'src'
							],
						}
					}
				}),
				compilerOptions: {
					// enable run-time checks when not in production
					dev: !production
				}
			}),
			// manually copy fontsource fonts, since rollup refuses to do it on its own
			copy({
				targets: [
					{ src: 'node_modules/@fontsource/*/files/*', dest: `${buildDir}/fonts` }
				]
			}),
			// we'll extract any component CSS out into
			// a separate file - better for performance
			css({ output: `${buildDir}/bundle-${type}.css` }),

			// If you have external dependencies installed from
			// npm, you'll most likely need these plugins. In
			// some cases you'll need additional configuration -
			// consult the documentation for details:
			// https://github.com/rollup/plugins/tree/master/packages/commonjs
			resolve({
				browser: true,
				dedupe: ['svelte']
			}),
			commonjs(),

			// If we're building for production (npm run build
			// instead of npm run dev), minify
			production && terser()
		],
		watch: {
			clearScreen: false
		}
	};
}

export default [
	themeGenerator('eager', undefined),
	themeGenerator('light'),
	themeGenerator('dark'),
	bundleGenerator('eager', 'appEager'),
	bundleGenerator('deferred', 'app', 'window'),
];
