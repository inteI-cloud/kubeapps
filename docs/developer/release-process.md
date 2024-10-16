# Kubeapps Releases Developer Guide

The purpose of this document is to guide you through the process of releasing a new version of Kubeapps.

## 0 - Ensure all 3rd-party dependencies are up to date

This step aims at decreasing the number of outdated dependencies so that we can get the latest patches with bug and security fixes.
It consists of four main stages: update the development images, update the CI, update the chart and, finally, update the dependencies.

### 0.1 - Development images

For building the [development container images](https://hub.docker.com/u/kubeapps), a number of base images are used in the build stage Specifically:

- The [dashboard/Dockerfile](../../dashboard/Dockerfile) uses:
  - [bitnami/node](https://hub.docker.com/r/bitnami/node/tags) for building the static files for production.
  - [bitnami/nginx](https://hub.docker.com/r/bitnami/nginx/tags) for serving the HTML and JS files as a simple web server.
- Those services written in Golang use the same image for building the binary, but then a [scratch](https://hub.docker.com/_/scratch) image is used for actually running it. These Dockerfiles are:
  - [apprepository-controller/Dockerfile](../../cmd/apprepository-controller/Dockerfile).
  - [asset-syncer/Dockerfile](../../cmd/asset-syncer/Dockerfile).
  - [assetsvc/Dockerfile](../../cmd/assetsvc/Dockerfile).
  - [kubeops/Dockerfile](../../cmd/kubeops/Dockerfile).
- The [pinniped-proxy/Dockerfile](../../cmd/pinniped-proxy/Dockerfile) uses:
  - [\_/rust](https://hub.docker.com/_/rust) for building the binary.
  - [bitnami/minideb:buster](https://hub.docker.com/r/bitnami/minideb) for running it.

> As part of this release process, these image tags _must_ be updated to the latest minor/patch version. In case of a major version, the change _should_ be tracked in a separate PR.

> **Note**: as the official container images are those being created by Bitnami, we _should_ ensure that we are using the same major version as they are using.

### 0.2 - CI configuration and images

In order to be in sync with the container images during the execution of the different CI jobs, it is necessary to also update the CI image versions.
Find further information in the [CI configuration](./ci.md) and the [e2e tests documentation](./end-to-end-tests.md).

#### 0.2.1 - CI configuration

In the [CircleCI configuration](../../.circleci/config.yml) we have an initial declaration of the variables used along with the file.
The versions used there _must_ match the ones used for building the container images. Consequently, these variables _must_ be changed accordingly:

- `GOLANG_VERSION` _must_ match the versions used by our services written in Golang, for instance, [kubeops](../../cmd/kubeops/Dockerfile).
- `NODE_VERSION` _must_ match the **major** version used by the [dashboard](../../dashboard/Dockerfile).
- `RUST_VERSION` _must_ match the version used by the [pinniped-proxy](../../dashboard/Dockerfile).
- `POSTGRESQL_VERSION` _must_ match the version used by the [Bitnami PostgreSQL chart](https://github.com/bitnami/charts/blob/master/bitnami/postgresql/values.yaml).

> As part of this release process, these variables _must_ be updated accordingly. Other variable changes _should_ be tracked in a separate PR.

#### 0.2.2 - CI integration image

- The [integration/Dockerfile](../../integration/Dockerfile) uses a [bitnami/node](https://hub.docker.com/r/bitnami/node/tags) image for running the e2e test.

> As part of this release process, this image tag _may_ be updated to the latest minor/patch version. In case of a major version, the change _should_ be tracked in a separate PR.

> **Note**: this image is not being built automatically. Consequently, a [manual build process](./end-to-end-tests.md#building-the-"kubeapps/integration-tests"-image) _must_ be triggered if you happen to upgrade the integration image.

### 0.3 - Development chart

Even though the official [Bitnami chart](https://github.com/bitnami/charts/tree/master/bitnami/kubeapps) is automatically able to retrieve the latest dependency versions, we still need to sync the versions declared in our own development chart.

#### 0.3.1 - Chart images

Currenty, the [values.yaml](../../chart/kubeapps/values.yaml) uses the following container images:

- [bitnami/nginx](https://hub.docker.com/r/bitnami/nginx/tags)
- [bitnami/kubectl](https://hub.docker.com/r/bitnami/kubectl/tags)
- [bitnami/oauth2-proxy](https://hub.docker.com/r/bitnami/oauth2-proxy/tags)

> As part of this release process, these image tags _must_ be updated to the latest minor version. In case of a major version, the change _should_ be tracked in a separate PR.

#### 0.3.2 - Chart dependencies

The chart [requirements.yaml](../../chart/kubeapps/requirements.yaml) _must_ be checked to ensure the version includes the latest dependent charts.

- Check if the latest versions are already included by running:

```bash
helm dependency list ./chart/kubeapps
```

- If they are not, run this other command to update the `requirements.lock` file:

```bash
helm dependency update ./chart/kubeapps
```

> As part of this release process, the chart dependencies _must_ be updated to the latest versions. In case of a major version, the change _should_ be tracked in a separate PR.

### 0.4 - Upgrading the code dependencies

Currently, we have three types of dependencies: the [dashboard dependencies](../../dashboard/package.json), the [golang dependencies](../../go.mod), and the [rust dependencies](../../cmd/pinniped-proxy/Cargo.toml). They _must_ be upgraded to the latest minor/patch version to get the latest bug and security fixes.

- Upgrade the [dashboard dependencies](../../dashboard/package.json) by running:

```bash
cd dashboard
yarn upgrade
```

- Check the outdated [golang dependencies](../../go.mod) by running the following (from [How to upgrade and downgrade dependencies](https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies)):

```bash
go mod tidy
go list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' -m all 2> /dev/null
```

Then, try to manually update those versions that can be safely upgraded. A useful tool for doing so is [go-mod-upgrade](https://github.com/oligot/go-mod-upgrade).

- Upgrade the [rust dependencies](../../cmd/pinniped-proxy/Cargo.toml) by running:

```bash
cd cmd/pinniped-proxy/
cargo update
```

- Finally, look at the [pull requests](https://github.com/kubeapps/kubeapps/pulls) and ensure there is no PR open by Snyk fixing a security issue. If so, discuss it with another Kubeapps maintainer and come to a decision on it, trying not to release with a high/medium severity issue.

> As part of this release process, the dashboard deps _must_ be updated, the golang deps _should_ be updated, the rust deps _should_ be updated and the security check _must_ be performed.

## 1 - Send a PR to the bitnami/chart repository

Since the chart that we host in the Kubeapps repository is only intended for development purposes, we need to synchronize it with the official one in the [bitnami/charts repository](https://github.com/bitnami/charts/tree/master/bitnami/kubeapps). To this end, we need to send a PR with the changes to their repository and wait until it gets accepted. Please note that the changes in both charts may involve additions and deletions, so we need to handle them properly (e.g., deleting the files first, performing a rsync, etc.).

> This step is currently manual albeit prone to change shortly.

## 2 - Select the commit to tag and perform a manual test

Once the dependencies have been updated and the chart changes merged, the next step is to choose the proper commit so that we can base the release on it. It is, usually, the latest commit in the main branch.

Even though the existing test base in our repository, we still _should_ perform a manual review of the application as it is in the selected commit. To do so, follow these instructions:

- Perform a checkout of the chosen commit.
- Install Kubeapps using the development chart: `helm install kubeapps ./chart/kubeapps/ -n kubeapps`
  - Note that if you are not using the latest commit in the main branch, you may have to locally build the container images so that the cluster uses the proper images.
- Ensure the core functionality is working:
  - Add a repository
  - Install an application from the catalog
  - Upgrade this application
  - Delete this application
  - Deploy an application in an additional cluster

## 3 - Create a git tag

Next, create a tag for the aforementioned commit and push it to the main branch. Please note that the tag name will be used as the release name.

For doing so, execute the following commands:

```bash
export VERSION_NAME="v1.0.0-beta.1" # edit it accordingly

git tag ${VERSION_NAME}
git push origin ${VERSION_NAME}
```

A new tag pushed to the repository will trigger, apart from the usual test and build steps, a _release_ [workflow](https://circleci.com/gh/kubeapps/workflows) as described in the [CI documentation](./ci.md).

> When a new tag is detected, Bitnami will automatically build a set of container images based on the tagged commit. They later will be published in [the Dockerhub image registry](https://hub.docker.com/search?q=bitnami%2Fkubeapps&type=image).

## 4 - Complete the GitHub release notes

Once the release job is finished, you will have a pre-populated [draft GitHub release](https://github.com/kubeapps/kubeapps/releases).

You still _must_ add a high-level description with the release highlights. Please take apart those commits just bumping dependencies up; it may prevent important commits from being clearly identified by our users.

Then, save the draft and **do not publish it yet** and get these notes reviewed by another Kubeapps maintainer.

## 5 - Publish the GitHub release

Once the new version of the [Kubeapps official chart](<(https://github.com/bitnami/charts/tree/master/bitnami/kubeapps)>) has been published and the release notes reviewed, you are ready to publish the release by clicking on the _publish_ button in the [GitHub releases page](https://github.com/kubeapps/kubeapps/releases).

> Take into account that the chart version will be eventually published as part of the usual Bitnami release cycle. So expect this step to take a certain amount of time.

## 6 - Promote the release

Tell the community about the new release by using our Kubernetes slack [#kubeapps channel](https://kubernetes.slack.com/messages/kubeapps). If it includes major features, you might consider promoting it on social media.
