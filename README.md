# Custom-plugin-demo

* RBAC
* Cache

### build plugin extension

```sh
go build -buildmode=plugin -o extName.so [...needed files]
```

if you see the error as following from log: 
```bash
time="2019-03-06T14:23:03+08:00" level=error msg="plugin.InstallExtension() got error: plugin.Open(\"cache\"): plugin was built with a different version of package github.com/jademperor/api-proxier/internal/logger, skip this" file="engine/engine.go:70"
``` 
you must make sure the `api-proxier` is same version with `your-plugin` depends on. the best way is recompiling two file. Make sure: **github.com/jademperor/api-proxier/plugin**'s version is correct. like this:
```bash
# go1.11+ get newest master version of package
go get github.com/jademperor/api-proxier@master

# or in your go.mod file
replace github.com/jademperor/api-proxier => path/to/api-proxier
```

### Usage

apiproxier -plugin=extName:ext.so:config.extName.json -plugin=extName:ext.so:config.extName.json ...

### Notices

* must implement **[github.com/jademperor/api-proxier/plugin.Plugin](https://github.com/jademperor/api-proxier/blob/master/plugin/plugin.go#L27)**, prototype if following:

```go
// Plugin type Plugin want to save all plugin
type Plugin interface {
	Handle(ctx *Context)
	Status() PlgStatus
	Enabled() bool
	Name() string
	Enable(enabled bool)
}
```

* must have **[New]** func, prototype if following:

```go
func New(cfgData []byte) plugin.Plugin
```