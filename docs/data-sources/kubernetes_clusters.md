---
page_title: "emma_kubernetes_clusters Data Source - emma"
subcategory: ""
description: |-
  Returns a list of Kubernetes clusters in the project, with optional filtering by project ID.
---

# emma_kubernetes_clusters (Data Source)

Returns a list of Kubernetes clusters managed by the Emma platform. Use this data source to enumerate existing clusters, check their status and version, or reference cluster details when configuring workloads.

All filter attributes are optional — omit `project_id` to return clusters across all projects your credentials can access.

## Example Usage

### List all Kubernetes clusters

```terraform
data "emma_kubernetes_clusters" "all" {}

output "cluster_names" {
  value = [for c in data.emma_kubernetes_clusters.all.kubernetes_clusters : c.name]
}
```

### List clusters in a specific project

```terraform
data "emma_kubernetes_clusters" "project" {
  project_id = 1297
}

output "active_clusters" {
  value = [
    for c in data.emma_kubernetes_clusters.project.kubernetes_clusters : c
    if c.status == "active"
  ]
}
```

### Find clusters by connection type and Kubernetes version

```terraform
data "emma_kubernetes_clusters" "all" {}

output "public_v128_clusters" {
  value = [
    for c in data.emma_kubernetes_clusters.all.kubernetes_clusters : c
    if c.k8s_connection_type == "public" && startswith(c.version, "1.28")
  ]
}
```

## Schema

### Optional

- `project_id` (Number) — Filter Kubernetes clusters by project ID. Omit to return clusters across all accessible projects.

### Read-Only

- `kubernetes_clusters` (List of Object) — List of matching Kubernetes clusters. Each element has:
  - `id` (Number) — Kubernetes cluster ID
  - `name` (String) — Kubernetes cluster name
  - `status` (String) — Status of the Kubernetes cluster (e.g. "active", "creating", "deleting")
  - `control_plane_status` (String) — Control plane status (e.g. "running", "unavailable")
  - `version` (String) — Kubernetes version (e.g. "1.28.5")
  - `deployment_location` (String) — Deployment region of the Kubernetes cluster
  - `k8s_connection_type` (String) — Kubernetes cluster network connectivity type (e.g. "public", "private")
  - `domain_name` (String) — Domain name attached to the Kubernetes cluster
  - `created_at` (String) — Date and time of cluster creation (ISO 8601)
  - `created_by_name` (String) — Name of the user who created the cluster
  - `created_by_id` (Number) — ID of the user who created the cluster
  - `modified_at` (String) — Date and time when the cluster was last modified (ISO 8601)
  - `modified_by_name` (String) — Name of the user who last modified the cluster
  - `modified_by_id` (Number) — ID of the user who last modified the cluster
