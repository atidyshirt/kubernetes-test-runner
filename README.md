# In Cluster Testing Kit

This guide outlines how to manage and test the WireMock example within a local [kind](kind.sigs.k8s.io) cluster using [DevSpace](devspace.sh).

## 🚀 Workflows

### 1. Prerequisites

Ensure you have the following installed:

* **Docker**
* [kind](kind.sigs.k8s.iodocs/user/quick-start/)
* [DevSpace CLI](devspace.shdocs/getting-started/installation)

### 2. Initial Setup
Create your local cluster and prepare the namespace:

```bash
kind create cluster --name k8s-test-runner
```

### 3. Workflow A: On-Demand Testing

Use this for one-off verification of your current code.

```bash
# 1. Deploy manifests and sync files
devspace deploy -n test
devspace sync -n test

# 2. Run the test suite once inside the pod
devspace run test
```

### 4. Workflow B: Interactive Development (Watch Mode)

> NOTE: Not currently functioning

Use this for a continuous feedback loop where tests re-run automatically on save.

```bash
devspace dev -n test
devspace run watch
```
