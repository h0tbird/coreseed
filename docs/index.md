---
title: The catabolic catalyst
---

<div class="jumbotron text-center">
  <h1>Káto</h1>
  <p class="lead">
    The magic is underneath
  </p>
  <p>
    <a href="https://github.com/katosys/kato/releases"
        class="btn btn-lg btn-primary">
      Káto releases
    </a>
  </p>
</div>


[Káto](https://github.com/katosys/kato) (from Greek *κάτω*: 'down', 'below', 'beneath', 'underneath') is an opinionated software-defined infrastructure (*SDI*) designed to control diverse computing workloads and work-flows.
Like in catabolism (from Greek *καταβολή*: 'throwing down'), the *Káto* system is the catalyst used to breakdown complex monolithic platforms into its fundamental microservices.

<img src="{{ site.baseurl}}/img/kato.png" width="70%" height="70%" align="right">

<br>

**Distinctive attributes**

- Geolocation
- Multidatacenter
- Cloud agnostic
- Variable costs
- Hardware abstraction
- Endo/exo-elasticity
- Microservices
- Containerization
- Task scheduling
- CI/CD pipelines
- Service discovery
- Load balancing
- High availability
- Self-healing

<br><br><br>

# Overview

[CoreOS](https://coreos.com/) is the foundation on which *Káto* is built. It provides the fundamental components used to assemble container-based distributed systems: [etcd](https://github.com/coreos/etcd) is used for consensus and discovery, [flannel](https://github.com/coreos/flannel) is used for virtual networking and [rkt](https://github.com/coreos/rkt) and [docker](https://github.com/docker/docker) are container engines.

All this *CoreOS* goodies are used to bootstrap a [Mesos](https://github.com/apache/mesos) cluster. *Mesos* is a distributed systems kernel which abstracts compute resources away from machines. Accordingly, it provides schedulers (or frameworks in *Mesos* parlance) which can run on top in order to utilise the exposed compute resources.

[Marathon](https://github.com/mesosphere/marathon) is one of such frameworks. It is a cluster-wide init and control system for long-running applications. Other frameworks like [Jenkins](https://github.com/jenkinsci/mesos-plugin) and [Elasticsearch ](https://github.com/mesos/elasticsearch) might coexist and share the same cluster resources.

[REX-Ray](http://rexray.readthedocs.io/en/stable/) delivers persistent storage access for container runtimes, such as *Docker* and *Mesos*, and provides an easy interface for enabling advanced storage functionality across common storage, virtualization and cloud platforms.
