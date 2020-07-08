@ECHO OFF
FOR /F "USEBACKQ tokens=*" %%F IN (`git rev-parse HEAD`) DO SET head=%%F
FOR /F "USEBACKQ tokens=*" %%F IN (`git describe --tags`) DO SET tags=%%F
go install -ldflags="-X github.com/pegnet/pegnetd/config.CompiledInBuild=%head% -X github.com/pegnet/pegnetd/config.CompiledInVersion=%tags%"