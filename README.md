# K8S-CM-To-Service

The k8s-cm-to-service is a little controller to bridge the gap between usual applications, connecting to kubernetes-services and the information created by [AWS K8s Controllers](https://github.com/aws-controllers-k8s/). 

The [ACK-Controllers](https://github.com/aws-controllers-k8s/community) allow to manage AWS-Services like [RDS](https://github.com/aws-controllers-k8s/rds-controller) through Kubernetes-Resources. In order to provide the necessary connection information, those controllers can update a given config-map through [FieldExport](https://github.com/aws-controllers-k8s/runtime/blob/main/apis/core/v1alpha1/field_export.go).
Since those generated config-maps do not necessarily fit into the configuration mechanisms of a given app, this controller observes such configmaps(in case they have a well-known label) and creates [Services](https://kubernetes.io/docs/concepts/services-networking/service/) out of them. 

## Configuration

TODO

## Example

TODO