# Plugin Schema Generator Tool

This tool generates a `plugin.yml` schema for a Buildkite plugin's configuration types using `invopop/jsonschema`. It
reflects the Go structs representing the plugin's configuration and outputs a YAML schema that can be used for
validation against what Buildkite passes to the plugin.

For more details on the `plugin.yml` specification, see the [Buildkite
documentation](https://buildkite.com/docs/pipelines/integrations/plugins/writing#step-2-add-a-plugin-dot-yml).

## 📖 Usage

### **Generate a Schema (Default Output)**

```bash
go run ./tools/schema/
```

### **Help Command**

```bash
go run ./tools/schema/ --help
```

## ✅ Example Output (`plugin.yml`)

```yaml
name: Terraform
description: A Buildkite plugin for processing terraform plans & applies,
with additional support for annotations & OPA checks against plans.
author: https://github.com/xphir
requirements: []
configuration:
  type: object
  properties:
    environments:
      type: array
      items:
        type: object
        properties:
          name:
            type: string
          clusters:
            type: array
  required:
    - environments
    - templates
```
