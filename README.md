# Dataverse Broker

[![Build Status](https://travis-ci.org/dataverse-broker/dataverse-broker.svg?branch=master)](https://travis-ci.org/dataverse-broker/dataverse-broker "Travis")

A go service broker for [Dataverse](https://dataverse.org) that implements the
[Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker).

This project is an implementation of [`osb-starter-pack`](https://github.com/pmorie/osb-starter-pack).

## Who should use this project?

You should use this project if you're interfacing a containerized application in Kubernetes that will utilize data stored on Dataverse.

## Prerequisites

You'll need:

- [`go`](https://golang.org/dl/) programming language
- A running [Kubernetes](https://github.com/kubernetes/kubernetes) (or [openshift](https://github.com/openshift/origin/)) cluster
- The [service-catalog](https://github.com/kubernetes-incubator/service-catalog)
  [installed](https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md)
  in that cluster

If you're using [Helm](https://helm.sh) to deploy this project, you'll need to
have it [installed](https://docs.helm.sh/using_helm/#quickstart) in the cluster.
Make sure [RBAC is correctly configured](https://docs.helm.sh/using_helm/#rbac)
for helm.

## Getting started

You can `go get` or `git clone` this repo.

### Get the project

```console
$ go get github.com/dataverse-broker/dataverse-broker/cmd/dataverse-broker
```

Or clone the repo:

```console
$ cd $GOPATH/src && mkdir -p github.com/dataverse-broker && cd github.com/dataverse-broker && git clone git://github.com/dataverse-broker/dataverse-broker
```

Change into the project directory:

```console
$ cd $GOPATH/src/github.com/dataverse-broker/dataverse-broker
```

### Deploy broker using Helm

```console
$ make deploy-helm
```

### Deploy broker using Openshift

```console
$ make deploy-openshift
```

Running either of these flavors of deploy targets will build the dataverse-broker binary,
build the image, deploy the broker into your Kubernetes, and add a
`ClusterServiceBroker` to the service-catalog.

## Using a Dataverse Service

### Using the Catalog

When logging in, if you are not automatically directed to the service catalog, you can do so manually by using the dropdown menu labelled "Add to Project" and selecting "Browse Catalog." There, you will see dataverse subtree icons among the list of services supported by the catalog.

### Utilizing a Service

To begin the process of provisioning and binding a dataverse subtree service, click on a dataverse subtree icon on the service catalog to generate a dialog window. The dialog window contains the following information in the order presented:

#### Information

![Information](/screenshots/Information.png?raw=true "Information tab of a Dataverse Service")

Provides a description of the corresponding dataverse subtree, including plans, if more than one. This tab is purely educational, and has no bearing on the actual provisioning/binding phase of the service.

#### Configuration

![Configuration](/screenshots/Configuration.png?raw=true "Configuration tab of a Dataverse Service")

Configure service to be provisioned/binded. Along with prompts to create a new project, you will be prompted to enter your API-token for this subtree (optional). The broker will check that your token has the necessary credentials to access that dataverse. During the Results tab, the provision step will fail if a provided token is invalid.

#### Service Binding

![Binding](/screenshots/Binding.png?raw=true "Binding tab of a Dataverse Service")

Allows for the option to bind the service and store the necessary information in a secret, or to create the binding at a later time inside a project.

#### Results

![Results](/screenshots/Results.png?raw=true "Results tab of a Dataverse Service")

At this page, the broker will attempt to provision and bind the service. Upon successful provision, the bind step will create a secret with the dataverse coordiates and your credentials. Use this secret with your created project to connect to the Dataverse Service.

### Add Secret to Application

For this section we'll asume that you're using the [`sample-dataverse-app`](https://github.com/dataverse-broker/sample-dataverse-app) as your application.

#### Provisioned Service in Project Overview

![Project Overview](/screenshots/ProjectOverview.png?raw=true "Project overview showing a provisioned service")

On your project page, you'll see your provisioned service. Expand the service by clicking on the arrow to the left of the name to see the secret you've created. Click "View Secret" to see the contents of the secret.

#### Viewing Your Secret

![Secret](/screenshots/Secret.png?raw=true "Secret")

The values of the secret parameters are hidden by default. You can view them by selecting "Reveal Secret".

The secret can be used inside your application. To add this secret to your application, click on the button on the top right labelled "Add to Application". This will open a window as illustrated in the figure below.

#### Adding the Secret to your Application

![Add to Application](/screenshots/Secret-AddToApplication.png?raw=true "Adding secret to sample application")

You'll have the option to add the secret in the form of environment variables or as a volume. For the sample application, select "Environment variables" as illustrated in the figure above. This will allow the application to have access to the coordinates and credentials in the secret. Click the blue "Save" button on the lower right of the window.

### Usage of whitelist

In order for a dataverse to be offered as a service, we need a bit of info regarding the specific dataverse in the form of metadata which is injected into an image (`json` object located in the whitelist folder residing in the image folder) which dataverse broker eventually calls upon in the event of a service binding. In the dataverse-broker/pkg/broker/utils.go file there are 2 functions of which get the metadata for a dataverse ( DataverseMetadataIds, and DataverseMeta ) which if you run the DataverseMetadataIds it will obtain the metadata for the dataverse. From there you use this output and create a `json` object similar to that of the current `json` objects in the whitelist folder, and you inject it with the output from the function. The "service_id" and "plan_id" fields are just UUIDs that can be generated online, unique for each service/plan, which is also injected into the `json` object for the dataverse.

## Goals of this project

- Make it easy for clients to interact with Dataverse
- Access datasets for use in containerized applications
