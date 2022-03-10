# pslive

Public livestreaming and real-time interactivity for Planktoscopes

## Usage

The binaries generated by this project are entirely self-contained, except for the other services which pslive needs to talk to. You can simply run the pslive binary from anywhere, though you'll need to set some environment variables to specify how pslive should talk to the services it depends on.

### Development

Before you start the server for the first time, you'll need to generate the webapp build artifacts by running `make buildweb`. Then you can start the server using golang's `go run` by setting some environment variables and running `make run`. You will need to have installed golang first. Any time you modify the webapp files (in the web/app directory), you'll need to run `make buildweb` again to rebuild the bundled CSS and JS. Whenever you use a CSS selector in a template file (in the web/templates directory), you should *also* run `make buildweb`, because the build process for the bundled CSS omits any selectors not used by the templates.

### Building

To execute the full build pipeline, run `make`; to build the docker images, run `make build`. You will need to have installed golang and golangci-lint first; `make build` *should* automatically install golangci-lint, but this might not work; if not, you'll need to install it manually. Note that `make build` will also automatically regenerate the webapp build artifacts. The resulting built binaries can be found in directories within the dist directory corresponding to OS and CPU architecture (e.g. `./dist/pslive_window_amd64/pslive.exe` or `./dist/pslive_linux_amd64/pslive`)

### Environment Variables

You'll need to set some environment variables to tell pslive how to assign names and how to connect to a ZeroTier network controller. Specifically, you'll need to set:

- INSTRUMENT_MJPEGSTREAM, which should be the URL of the MJPEG stream from the Planktoscope.
- SESSIONS_COOKIE_NOHTTPSONLY, which should be `true` if you are running pslive locally (as `localhost`) without HTTPS. If you are running pslive over the web, you should run it behind an HTTPS reverse proxy and you should leave SESSION_COOKIE_NOHTTPSONLY unset.
- SESSIONS_AUTH_KEY, which should be set to a session key generated by running pslive without the SESSION_AUTH_KEY set.
- AUTHN_ADMIN_PW_HASH, which should be set to the password hash generated by running pslive with a password set as AUTHN_ADMIN_PW.

For example, you could generate the password and session key using:
```
AUTHN_ADMIN_PW='mypassword' make run
```
which will print a message like:
```
Record this admin password hash for future use as AUTHN_ADMIN_PW_HASH
(use single-quotes from shell to avoid string substitution with dollar-signs):
$argon2id$v=19$m=65536,t=1,p=2$EIV/HJ0DILHeNf2IC+qsGQ$BvBCCEsKUCKuAPI+pzM+sbCy/pdQdOF/FmHwx/yIusU
Record this key for future use as SESSIONS_AUTH_KEY:
QVG4y5EPPoDZjAzYc6j7I09iJum3w+hXNrB3O4HQvSc=
```

And then you could run the server in development mode (which you can log into with username `admin` and password `mypassword`) using:
```
INSTRUMENT_MJPEGSTREAM='https://mjpeg-proxy.cloud.syngnathus.sargassum.world/' \
SESSION_AUTH_KEY='QVG4y5EPPoDZjAzYc6j7I09iJum3w+hXNrB3O4HQvSc=' \
AUTHN_ADMIN_PW_HASH='$argon2id$v=19$m=65536,t=1,p=2$EIV/HJ0DILHeNf2IC+qsGQ$BvBCCEsKUCKuAPI+pzM+sbCy/pdQdOF/FmHwx/yIusU' \
make run
```

Or you could run the built binary using:
```
INSTRUMENT_MJPEGSTREAM='https://mjpeg-proxy.cloud.syngnathus.sargassum.world/' \
SESSION_AUTH_KEY='QVG4y5EPPoDZjAzYc6j7I09iJum3w+hXNrB3O4HQvSc=' \
AUTHN_ADMIN_PW_HASH='$argon2id$v=19$m=65536,t=1,p=2$EIV/HJ0DILHeNf2IC+qsGQ$BvBCCEsKUCKuAPI+pzM+sbCy/pdQdOF/FmHwx/yIusU' \
./pslive
```

## License

Copyright Prakash Lab and the Sargassum project contributors.

SPDX-License-Identifier: Apache-2.0 OR BlueOak-1.0.0

You can use this project either under the [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0) or under the [Blue Oak Model License 1.0.0](https://blueoakcouncil.org/license/1.0.0); you get to decide. We chose the Apache license because it's OSI-approved, and because it goes well together with the [Solderpad Hardware License](http://solderpad.org/licenses/SHL-2.1/), which is a license for open hardware used in other related projects but not this project. We prefer the Blue Oak Model License because it's easier to read and understand.
