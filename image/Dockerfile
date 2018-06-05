FROM ubuntu:16.04

RUN mkdir -p /usr/lib/oracle/12.1/client64
ADD lib /usr/lib/oracle/12.1/client64/lib
ADD servicebroker /opt/servicebroker/servicebroker
ENV LD_LIBRARY_PATH=/usr/lib/oracle/12.1/client64/lib
CMD /opt/servicebroker/servicebroker --help
