# Tekton Exporter

<img src="https://raw.githubusercontent.com/achetronic/tekton-exporter/master/docs/img/logo.png" alt="Tekton Exporter Logo (Main) logo." width="150">

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/achetronic/tekton-exporter)
![GitHub](https://img.shields.io/github/license/achetronic/tekton-exporter)

![YouTube Channel Subscribers](https://img.shields.io/youtube/channel/subscribers/UCeSb3yfsPNNVr13YsYNvCAw?label=achetronic&link=http%3A%2F%2Fyoutube.com%2Fachetronic)
![X (formerly Twitter) Follow](https://img.shields.io/twitter/follow/achetronic?style=flat&logo=twitter&link=https%3A%2F%2Ftwitter.com%2Fachetronic)

A CLI to TODOOOOOOOO

## Motivation

TODO

## Flags

Every configuration parameter can be defined by flags that can be passed to the CLI.
They are described in the following table:

| Name              | Description                      | Default Example |                                 |
|:------------------|:---------------------------------|:---------------:|---------------------------------|
| `--log-level`     | Define the verbosity of the logs |     `info`      | `--log-level info`              |
| `--disable-trace` | Disable traces from logs         |     `false`     | `--disable-trace true`          |
| `--kubeconfig`    | Path to kubeconfig               |       `-`       | `--kubeconfig="~/.kube/config"` |

## Examples

Here you have a complete example to use this command.

> Output is thrown always in JSON as it is more suitable for automations

```console
tekton-exporter run \
    --log-level=info
    --kubeconfig="./path"
```

> ATTENTION:
> If you detect some mistake on the examples, open an issue to fix it. This way we all will benefit

## How to use

This project provides binary files and Docker images to make it easy to use wherever wanted

### Binaries

Binary files for the most popular platforms will be added to the [releases](https://github.com/achetronic/tekton-exporter/releases)

### Docker

Docker images can be found in GitHub's [packages](https://github.com/achetronic/tekton-exporter/pkgs/container/tekton-exporter)
related to this repository

> Do you need it in a different container registry? We think this is not needed, but if we're wrong, please, let's discuss
> it in the best place for that: an issue

## How to contribute

We are open to external collaborations for this project: improvements, bugfixes, whatever.

For doing it, open an issue to discuss the need of the changes, then:

- Fork the repository
- Make your changes to the code
- Open a PR and wait for review

The code will be reviewed and tested (always)

> We are developers and hate bad code. For that reason we ask you the highest quality
> on each line of code to improve this project on each iteration.

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Special mention

This project was done using IDEs from JetBrains. They helped us to develop faster, so we recommend them a lot! 🤓

<img src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png" alt="JetBrains Logo (Main) logo." width="150">