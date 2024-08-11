# Hitman

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/achetronic/hitman)
![GitHub](https://img.shields.io/github/license/achetronic/hitman)

![YouTube Channel Subscribers](https://img.shields.io/youtube/channel/subscribers/UCeSb3yfsPNNVr13YsYNvCAw?label=achetronic&link=http%3A%2F%2Fyoutube.com%2Fachetronic)
![X (formerly Twitter) Follow](https://img.shields.io/twitter/follow/achetronic?style=flat&logo=twitter&link=https%3A%2F%2Ftwitter.com%2Fachetronic)

A daemon for Kubernetes to kill target resources under user-defined templated conditions

## Motivation

In today's fast-paced environments, Kubernetes clusters often manage systems that dynamically create and destroy resources automatically. Examples of these are pipelines and cronjobs. 

However, these automated processes can sometimes get stuck, causing disruptions that affect the smooth operation of the entire system. Often, simply terminating some of these objects can restore normalcy. 

There is a need for a solution that empowers Kubernetes administrators to automate this cleanup process efficiently. 
This project exists to provide a robust agent for automating the deletion of potential stuck resources, 
ensuring your Kubernetes clusters run smoothly and reliably.

## Flags

As every configuration parameter can be defined in the config file, there are only few flags that can be defined.
They are described in the following table:

| Name              | Description                    |    Default    | Example                  |
|:------------------|:-------------------------------|:-------------:|:-------------------------|
| `--config`        | Path to the YAML config file   | `hitman.yaml` | `--config ./hitman.yaml` |
| `--log-level`     | Verbosity level for logs       |    `info`     | `--log-level info`       |
| `--disable-trace` | Disable showing traces in logs |    `info`     | `--log-level info`       |

> Output is thrown always in JSON as it is more suitable for automations

```console
hitman run \
    --log-level=info
    --config="./hitman.yaml"
```

## Examples

Here you have a complete example. More up-to-date one will always be maintained in 
`docs/prototypes` directory [here](./docs/prototypes)


```yaml
version: v1alpha1
kind: Hitman
metadata:
  name: killing-sample
spec:
  synchronization:
    time: 1m

  resources:

    - target:
        group: ""
        version: v1
        resource: pods

        # Select the resources by their name
        # Choose one of the following options
        name:
          matchRegex: ^(coredns-)
          #matchExact: "coredns-xxxxxxxxxx-yyyyy"
        
        # Select the namespace where the resources are located
        # Choose one of the following options
        namespace: 
          matchRegex: ^(kube-system)
          #matchExact: kube-system
        
      conditions:

      # Delete the resources when they are older than 10 minutes
      - key: |-
          {{/* Define some variables */}}
          {{- $maxAgeMinutes := 10 -}}

          {{- $nowTimestamp := (now | unixEpoch) -}}
          {{- $podStartTime := (toDate "2006-01-02T15:04:05Z07:00" .object.status.startTime) | unixEpoch -}}
          
          {{/* Calculate the age of the resource in minutes */}}
          {{- $minutedFromNow := int (round (div (sub $nowTimestamp $podStartTime) 60) 0) -}}
            
          {{/* Print true ONLY if the resource is older than 5 minutes */}}
          {{- printf "%v" (ge $minutedFromNow $maxAgeMinutes) -}}
        value: true

```

> ATTENTION:
> If you detect some mistake on the config, open an issue to fix it. This way we all will benefit

## How to deploy

This project is designed specially for Kubernetes, but also provides binary files 
and Docker images to make it easy to be deployed however wanted

### Binaries

Binary files for most popular platforms will be added to the [releases](https://github.com/achetronic/hitman/releases)

### Kubernetes

You can deploy `hitman` in Kubernetes using Helm as follows:

```console
helm repo add hitman https://achetronic.github.io/hitman/

helm upgrade --install --wait hitman \
  --namespace hitman \
  --create-namespace achetronic/hitman
```

> More information and Helm packages [here](https://achetronic.github.io/hitman/)


### Docker

Docker images can be found in GitHub's [packages](https://github.com/achetronic/hitman/pkgs/container/hitman) 
related to this repository

> Do you need it in a different container registry? I think this is not needed, but if I'm wrong, please, let's discuss 
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
