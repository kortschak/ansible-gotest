# ansible-gotest

This is an ansible config and binary module helper for executing `go test` on an ansible target host.

To use ansible-gotest you will need to build the module for the target arch and place the executable in the [library directory](https://docs.ansible.com/ansible/latest/dev_guide/developing_locally.html#adding-a-module-locally). For example for arm64:
```
GOARCH=arm64 go build -o library gotest.go
```

Then use `config.yml.tmpl` to make a `config.yml` and set up your other ansible setting.

```
---
goroot: /usr/local/go
repo: <repo-URL>
path: src/gonum.org/v1/gonum
key_file: <path-to-ssh-private-key-for-source-repo>
```

The values of attributes in config.yml:
- `goroot`: the GOROOT on the target host
- `repo`: the git URL of the origin you want to test
- `path`: the destination path of the repository
- `key_file`: the absolute path to the SSH private key for the key used to access the git repo

When it has been set up, you can then run a playbook with, for example for gonum testing.

```
ansible-playbook gonum.yml -e "branch=<target-branch> pkg=<package-path> count=<n>"
```

The `gonum.yml` file can be modified for other repositories. Each of `branch`, `pkg` and `count` are optional.

- `branch`: is the branch of the repository to be tested (default: `master`)
- `pkg`: is the repo-relative path to the package to be tested using `go test` path syntax (default: `...`)
- `count`: is the `-count` option passed to `go test` (default: unset)

