# Using xgo

xgo is a cross compile tool that works with cgo. Since we use a sqlite library that includes cgo, the normal cross compile method does not work. The xgo Dockerfile in this directory will accomplish the cross compile.

If you use the Dockerfile in this directory:

```
docker build --tag xgo-builder .
xgo -ldflags="-X github.com/pegnet/pegnetd/config.CompiledInBuild=`git rev-parse HEAD` -X github.com/pegnet/pegnetd/config.CompiledInVersion=`git describe --tags`" -image xgo-builder --targets=windows/amd64,darwin/amd64,linux/amd64 .
```

