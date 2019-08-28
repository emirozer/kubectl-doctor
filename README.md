# Kubernetes CLI Plugin - Doctor

This plugin is inspired from [brew](http://brew.sh/) doctor :) It will scan your currently `target`ed k8s cluster to see if there are anomalies or useful action points that it can report back to you.

This plugin does *not* change any state or configuration, it merely just scans and gathers information than reports back anomalies in yaml format.

![Demo](./docs/example.svg)

## Install
1. Download a binary from [releases](https://github.com/emirozer/kubectl-doctor/releases) that is compatible with your os/arch
2. Rename it to `kubectl-doctor`
3. Add it to your `PATH`

## Usage
When the plugin binary is found from `PATH` you can just execute it through `kubectl` CLI
```shell
kubectl doctor
```

## Current list of anomaly checks

* core component health (etcd cluster members, scheduler, controller-manager)
* orphan endpoints (endpoints with no ipv4 attached)
* persistent-volume available & unclaimed
* persistent-volume-claim in lost state
* k8s nodes that are not in ready state
* orphan replicasets (desired number of replicas are bigger than 0 but the available replicas are 0)
* leftover replicasets (desired number of replicas and the available # of replicas are 0)
* orphan deployments (desired number of replicas are bigger than 0 but the available replicas are 0)
* leftover deployments (desired number of replicas and the available # of replicas are 0)
* leftover cronjobs (last active date is more than a month)
