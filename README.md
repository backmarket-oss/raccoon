# raccoon

![raccoon logo](logo.png "Hey raccoon")

[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](code_of_conduct.md)

## Description

Raccoon is a **Command Line Interface** aiming to provide **ephemerality** in a **kubernetes environment**. Ephemerality is a concept of things being transitory, existing only briefly.


Ephemerality will help to:
- reduce hacks probability on a pod
- avoid memory leak pitfalls
- ensure to not rely on long-running workloads
- ensure that we can restart a pod whenever we want/need
- help enforcing immutability

**For now, raccoon is only capable of collecting kubernetes pods resources.**

### Strategies

As raccoon is primarly focused on deleting pods. Those pods can start at the same time, when a new deployment
happens, deleting pods with a same start time will result in a service unavailability.  
To avoid this pitfall, raccoon is able to delete pods based on a strategy.  

#### Randomized Delay
For now, there is only one strategy, called `randomizedDelay` which basically dispatch deletion over a certain time interval,
to prevent service unavailability.

## Build and install from source

### Prerequisite tools
* Docker daemon
* Git
* Go
* pre-commit (used to generate README file for the helm chart)

### Install from Github

Pre-requisites
```bash
# Install golangci-lint
## MacOS
brew install golangci-lint
## Linux & Windows
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.45.2

# Verify
golangci-lint --version

# Install pre-commit hook
pre-commit install
```

```bash
mkdir $HOME/src
cd $HOME/src
git clone https://github.com/backmarket-oss/raccoon.git
cd raccoon
make install
```

### Pull docker image
raccoon docker image is available at `ghcr.io/backmarket-oss/raccoon:latest`.

## Usage
Each raccoon's flag is available as an environment variable with `RACCOON_` as prefix.

### Helm repository
Raccoon helm repository is available at https://backmarket-oss.github.io/raccoon/, please follow 
these instructions to use it:

```
$ helm repo add raccoon https://backmarket-oss.github.io/raccoon

$ helm search repo raccoon
NAME                    CHART VERSION   APP VERSION     DESCRIPTION               
backmarket-oss/raccoon  1.0.0           1.0.0           Ephemerality in kubernetes
```

Here is the available commands in raccoon
### garbage
Used to run raccoon daemon and start marking and collecting k8s pods.  
`namespace` and `selector` are the two mandatory flags.

For now, we have only one strategy to collect resources, it is the randomized delay strategy.  
At each `--check-interval` and for each resource to collect we apply a `--randomized-delay` to avoid deleting all the resources in one shot. 

```
$ raccoon garbage

Run raccoon daemon

Usage:
  raccoon garbage [flags]

Flags:
      --check-interval int     Interval between two raccoon check (default 120)
      --dry-run                Test process without deletion
  -h, --help                   help for garbage
      --kube-location string   Connection mode to the kubernetes api (in or out) (default "in")
      --kubeconfig string      Path to KUBECONFIG file. Ignored if KUBECONFIG envvar is set (default "${HOME}/.kube/config")
  -n, --namespace string       Namespace to raccoon (required)
      --randomized-delay int   Delay the deletion by a randomly amount of time [value/2,value] (default 120)
  -s, --selector string        Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2) (required)
      --ttl string             Minimum age by which a pod will be deleted (default 24h0m0s)

Global Flags:
      --level string   set log level (default "info")
```

# About the project
## Getting involved and contributing
See [contribute](./docs/CONTRIBUTE.md).

## Why naming it raccoon?

> Since they are omnivores, berries, fruit, eggs, lizards, crustaceans, fish, wild birds, domestic poultry, and garbage scraps are their main source of food. This is why they eat "garbage". To them, it is just another source of food that is easily accessible.

They eat "garbage", that is what we do in this project. Collecting and deleting garbage resources.

## License

Raccoon is released under the Apache 2.0 license. See [LICENSE](LICENSE)
