# kn-diag

## Introduction

kn-diag is a tool to expose information and status of knative resources to end-user for debugging purpose. 

kn-diag is inspired by [Kperf](https://github.com/zhanggbj/kperf) and [knative inspect](https://github.com/nimakaviani/knative-inspect).  

## Design documents

The goolge doc for the design purpose of kn-diag: 
https://docs.google.com/document/d/1i-lJXVMr0iPymck61UAiQCwffVpoUtMU-EK5wkPxAKo/edit?usp=sharing

## Build
To build it,  run cmd as below:

```
go build -o kn-diag ./cmd/main.go
```

## Support Cmds:
* kn-diag service <ksvc-name> --namespace <namespace-name>
* kn-diag service <ksvc-name> --namespace <namespace-name> --verbose conditions
* kn-diag service <ksvc-name> --namespace <namespace-name> --verbose keyinfo
  
## Demo
https://youtu.be/PxD0yxb8B7c


