# Grabit

Grabit is a utility that helps the definition, downloading and integrity validation of
external assets accessible via HTTPS/HTTP.

The integrity of the assets is verified by storing the
[subresource integrity](https://en.wikipedia.org/wiki/Subresource_Integrity) of the
assets and validating it every time the assets are downloaded.

It's typically used as part of a build pipeline when external assets need to be used and their
integrity needs to be validated to guard against [supply chain attacks](https://en.wikipedia.org/wiki/Supply_chain_attack).

## Installation

```sh
go install github.com/cisco-open/grabit@latest
```

## Usage

Typically usage involves 3 steps:

- Definition of the assets
- Lock file committing
- Asset downloading

### Definition of the assets

Manually run Grabit to generate the lock file `grabit.lock` with the definition of all the assets that will be
used during the asset downloading step:

```sh
$ grabit add https://example.com/
$ cat grabit.lock
[[Resource]]
Urls = ['https://example.com/']
Integrity = 'sha256-6o+sfGX7WJsNU1YPUlH3T56bJDR43Laz6nm142RJyNk='
```

### Lock file committing

The `grabit.lock` contains the list of all the assets defined in the previous step along with the information needed
to perform validation. You will want to commit this file in your source code repository.

### Asset downloading

The build pipeline will then consume the lock file by running the following to download all the assets and check
their integrity:

```sh
$ grabit download --dir .
# Use the assets...
```

## Support

We are continuously improving the tool and adding more feature.
Please see the [open issues](https://github.com/cisco-open/grabit/issues) to see the list of planned
items and feel free to open a new issue in case something that you'd like to see is missing.
