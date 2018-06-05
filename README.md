# Introduction
This is a sample service broker built using [osb-starter-kit](https://github.com/pmorie/osb-starter-pack)

The sample service broker enables creation of a user and table in Oracle DB. This is to demonstrate the ease
of building service brokers and how it can be used for integration with existing services in your datacenter

## Pre-req
You'll need to download Oracle Instant Client libraries for the OS you are building 

## Build Instructions
### Example instructions for Ubuntu 16.04 Power (ppc64le)
- Download and unzip the client library to /usr/lib/oracle/12.1/client64/lib
- Set LD_LIBRARY_PATH `export LD_LIBRARY_PATH=/usr/lib/oracle/12.1/client64/lib`
- Copy the libraries to images/lib
- Build image `IMAGE=bpradipt/servicebroker TAG=latest make push`

## Deployment
Deploy using the provided helm chart


# Authors
Pradipta Banerjee (bpradipt@in.ibm.com) <br>
Abhishek Dasgupta (abdasgupta@in.ibm.com)

