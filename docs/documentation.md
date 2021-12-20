# Documentation
The key components of _FlashPipe_ are
- **Java executables** - provide native access to SAP Integration Suite APIs
- **Unix scripts** - provide executable steps for configuration in CI/CD pipeline
- **Local Maven repository** - provides cached libraries for faster execution of Maven-based unit testing _(not available in lite version of Docker images)_

_FlashPipe_ uses the [public APIs of the SAP Integration Suite](https://api.sap.com/package/CloudIntegrationAPI?section=Artifacts) to automate the Build-To-Deploy cycle. The components are implemented in Groovy and compiled as Java executables.

While it is possible to use the Java executables directly, the Unix scripts do most of the heavy lifting by orchestrating between the various API calls required to complete the Build-To-Deploy cycle.

## Prerequisite
To use _FlashPipe_, you will need the following
1. Access to **Cloud Integration** on an SAP Integration Suite tenant - typically an Integration Developer credentials are required
2. Access to a **CI/CD platform**, e.g. [Azure Pipelines](https://azure.microsoft.com/en-us/services/devops/pipelines/), [GitHub Actions](https://github.com/features/actions)
3. **Git-based repository** to host the contents of the Cloud Integration artifacts

Technically, it should be possible to use _FlashPipe_ on any CI/CD platform that supports container-based pipeline execution and Unix script execution.

## Docker image tags
_FlashPipe_'s Docker images comes in two flavours. The difference between the full and lite tags for each version release is the inclusion of Maven capabilities in the image.
- **Full** (e.g. tag `2.4.6`)
  - The full image includes Maven and selected local repositories of third-party libraries. This can be used for Maven-based testing and the cached libraries improves the execution time.

- **Lite** (e.g. tag `2.4.6-lite`)
  - The lite image only contains the required third-party libraries (without the full-blown Maven) for execution of the Unix scripts. The smaller size reduces the time required to pull the image from Docker and is recommended when Maven is not used.

### Rolling tags
Starting from version `2.3.0`, rolling tags are introduced to make it easier to get the latest version. Rolling tags are dynamic and will point to the latest version of the corresponding image. The following rolling tags are available:
- `2.x.x` & `2.x.x-lite` - points to the latest release of major version 2
- `2.3.x` & `2.3.x-lite` - points to the latest release of minor version 2.3

### Usage recommendation
- When using _FlashPipe_ in productive pipelines, use an immutable tag (e.g. `2.3.0`) to ensure stability so that the pipeline will not be affected negatively by new versions.
- When using _FlashPipe_ in development pipelines, use a suitable (major/minor version) rolling tag to always get the latest version.

## Authentication
_FlashPipe_ supports the following methods of authentication when accessing the SAP Integration Suite APIs.
- Basic authentication
- OAuth authentication

It is recommended to use OAuth so that the access is not linked to an individual's credential (which may be revoked or the password might change). For details on setting up an OAuth client for use with _FlashPipe_, visit the [OAuth client setup page](oauth_client.md).

## Usage of Unix scripts
For details on usage of the Unix scripts in pipeline steps, visit the [Unix scripts page](unix-scripts.md).

## Usage examples
Following are different usage examples of _FlashPipe_ on different CI/CD platforms.
- [Upload/Deploy Integration Flows using Azure Pipelines](azure-pipelines-upload.md)
- [Upload/Deploy Integration Flows using GitHub Actions](github-actions-upload.md)
- [Sync Integration Flows from Tenant to GitHub using GitHub Actions](github-actions-sync.md)
- [Snapshot Tenant Content to GitHub using GitHub Actions](github-actions-snapshot.md)
- [Simulation Testing using Maven](simulation-testing.md)

## Reference
The following repository on GitHub provides sample usage of _FlashPipe_.

[https://github.com/engswee/flashpipe-demo](https://github.com/engswee/flashpipe-demo)