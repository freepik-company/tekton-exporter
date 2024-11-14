# Tekton Exporter

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/freepik-company/tekton-exporter)
![GitHub](https://img.shields.io/github/license/freepik-company/tekton-exporter)

![YouTube Channel Subscribers](https://img.shields.io/youtube/channel/subscribers/UCeSb3yfsPNNVr13YsYNvCAw?label=achetronic&link=http%3A%2F%2Fyoutube.com%2Fachetronic)
![GitHub followers](https://img.shields.io/github/followers/achetronic?label=achetronic&link=http%3A%2F%2Fgithub.com%2Fachetronic)
![X (formerly Twitter) Follow](https://img.shields.io/twitter/follow/achetronic?style=flat&logo=twitter&link=https%3A%2F%2Ftwitter.com%2Fachetronic)

A Prometheus Exporter to collect and expose some useful fine-grained Tekton metrics
that are not exposed natively

## Motivation

In our platform team, we needed to build some useful dashboards to show our developers the status of the pipelines
per project to make them compete for the first position on the reliability podium.

There are some metrics that are not exposed, which are necessary to achieve our goal,
so we decided to create this small tool.

## Flags

Every configuration parameter can be defined by flags that can be passed to the CLI.
They are described in the following table:

| Name                 | Description                                                             | Default Example |                                                            |
|:---------------------|:------------------------------------------------------------------------|:---------------:|------------------------------------------------------------|
| `--log-level`        | Define the verbosity of the logs                                        |     `info`      | `--log-level info`                                         |
| `--disable-trace`    | Disable traces from logs                                                |     `false`     | `--disable-trace true`                                     |
| `--kubeconfig`       | Path to kubeconfig                                                      |       `-`       | `--kubeconfig="~/.kube/config"`                            |
| `--metrics-port`     | Port where metrics web-server will run                                  |     `2112`      | `--metrics-port 9090`                                      |
| `--metrics-host`     | Host where metrics web-server will run                                  |    `0.0.0.0`    | `--metrics-host 10.10.10.1`                                |
| `--populated-labels` | (Repeatable or comma-separated list) Object labels populated on metrics |       `-`       | `--populated-labels "apiVersion,pipelineName,projectName"` |

> For Prometheus SDK, it is mandatory to register the metrics before using them.
> Due to this, if you use `--populated-labels` flag and the label is not present in some PipelineRun or TaskRun
> the label will be populated with `#` as value

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

Binary files for the most popular platforms will be added to the [releases](https://github.com/freepik-company/tekton-exporter/releases)

### Docker

Docker images can be found in GitHub's [packages](https://github.com/freepik-company/tekton-exporter/pkgs/container/tekton-exporter)
related to this repository

> Do you need it in a different container registry? We think this is not needed, but if we're wrong, please, let's discuss
> it in the best place for that: an issue

## Exposed metrics

This project is about exposing useful metrics related to the status of the Pipelines and Tasks, so, what about them?


| Name                                           | Description                     |                         Metric labels                          |
|:-----------------------------------------------|:--------------------------------|:--------------------------------------------------------------:|
| `tekton_exporter_pipelinerun_status`           | Status of a PipelineRun         |            `name`, `namespace`, `status`, `reason`             |
| `tekton_exporter_taskrun_status`               | Status of a TaskRun             |            `name`, `namespace`, `status`, `reason`             |
| `tekton_exporter_pipelinerun_duration_seconds` | Seconds lasted by a PipelineRun | `name`, `namespace`, `start_timestamp`, `completion_timestamp` |
| `tekton_exporter_taskrun_duration_seconds`     | Seconds lasted by a TaskRun     | `name`, `namespace`, `start_timestamp`, `completion_timestamp` |

## Deployment

We have designed the deployment of this project to allow remote deployment using Helm. This way it is possible
to use it with a GitOps approach, using tools such as ArgoCD or FluxCD.


> 🧚🏼 **Hey, listen! Go to the [Helm registry](https://freepik-company.github.io/tekton-exporter/)**

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
