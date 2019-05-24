# Deploying tensorio-models to Kubernetes clusters using helm

This helm chart allows for the deployment of any number of `tensorio-models` components
([`repository`](../cmd/repository), [`flea`](../cmd/flea), [aggregators](../aggregator), etc.) to
Kubernetes clusters.

To do so, all you need is to:

1. Create an appropriate [values file](https://github.com/helm/helm/blob/master/docs/chart_template_guide/values_files.md)
(and store it at `VALUES_FILEPATH`)

2. Make sure that the required secrets and configurations are present on your cluster of choice
(in the desired namespace)

3. Make sure you are pointed at the right kubernetes context

4. Determine a name for your release (and store it at `RELEASE_NAME`)

Then, from the project root, run:
```
helm upgrade --install $RELEASE_NAME -f $VALUES_FILEPATH helm/
```

This directory contains some example of values files, following the pattern `values.<purpose>.yaml`.
The `<purpose>` signifies what the corresponding file demonstrates.


## Functionality

In this section, we expand on some of the deployment functionality this helm chart offers.

### Repository

[`values.gcs.yaml`](./values.gcs.yaml) shows how to deploy the `tensorio-models`
[`repository`](../cmd/repository) APIs with a GCS backend.

[`values.basic.yaml`](./values.basic.yaml) shows how to make the same deployment with a memory
backend (for testing purposes).

### Flea

### Aggregators

To trigger an aggregation job on a given kubernetes cluster, use helm from project root:
```
helm upgrade --install \
    flea-aggregator \
    --set aggregator.tag=$DOCKER_TAG \
    --set aggregator.aggregationId=aggregation-$(date -u +%Y%m%d-%H%M) \
    --set aggregator.checkpointsFilePath=$CHECKPOINTS_FILEPATH \
    --set aggregator.outputPath=$OUTPUT_PATH \
    -f values.aggregator.yaml \
    helm/
```

Here, `values.aggregator.yaml` should look like `helm/values.aggregator.yaml` with perhaps modified
`aggregator.repository`.

### Secrets

Certain configurations of `tensorio-models` backend components will require access to secrets (for
example, Google Cloud Platform service account credentials). As such, these secrets **should not**
be commited into code or even into docker images. The graceful way to handle such secrets is to make
them available as [Kubernetes secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
and load them into containers at runtime (either by loading them in as environment variables or
mounting them on the container file system).

The existence of Kubernetes secrets is a dependency for the deployment that can only be resolved at
the time of release, and the default behaviour of `helm` is to complete releases even if pods that
were created as part of that release are failing.

You may prefer that the release itself fails if any required secrets are not present on the cluster.
To this end, this helm chart allows you to specify a `checkSecrets` directive in your values file.
This directive runs a pre-install check for the existence of any secrets specified as part of
`checkSecrets`. For an example of how to specify `checkSecrets`, see
[values.checkSecrets.yaml](./values.checkSecrets.yaml).
