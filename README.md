## Clean up unused iOS simulators

This step helps free up disk space by removing unused iOS simulators in Bitrise VMs.This step is supposed to run only on Xcode stacks.

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://devcenter.bitrise.io/steps-and-workflows/steps-and-workflows-index/).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

### Example

```yaml
steps:
- git::https://github.com/bitrise-silver/simulator-cleanup.git:
    inputs:
    - remove_versions_lower_than: '15.0'
```

## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `remove_versions_lower_than` | Remove simulators with lower-bound versions | | `0.0` |
| `remove_versions_higher_than` | Remove simulators with upper-bound versions | | `99.0` |
| `platform` | Remove simulators with specified platform | | `iOS` |
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-silver/simulator-cleanup/pulls) and [issues](https://github.com/bitrise-silver/simulator-cleanup/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://devcenter.bitrise.io/bitrise-cli/run-your-first-build/).

**Note:** this step may remove iOS simulators in your machine when run locally so please choose the parameter values carefully.

Learn more about developing steps:

- [Create your own step](https://devcenter.bitrise.io/contributors/create-your-own-step/)
- [Testing your Step](https://devcenter.bitrise.io/contributors/testing-and-versioning-your-steps/)