FROM nginx:alpine
# Stress Version can be found on offcial website of stress
# https://people.seas.harvard.edu/~apw/stress/
ENV STRESS_VERSION=1.0.4 \
    SHELL=/bin/bash

RUN \
  apk add --update bash g++ make curl nginx && \
  curl -o /tmp/stress-${STRESS_VERSION}.tgz https://people.seas.harvard.edu/~apw/stress/stress-${STRESS_VERSION}.tar.gz && \
  cd /tmp && tar xvf stress-${STRESS_VERSION}.tgz && rm /tmp/stress-${STRESS_VERSION}.tgz && \
  cd /tmp/stress-${STRESS_VERSION} && \
  ./configure && make && make install && \
  apk del g++ make curl && rm -rf /tmp/* /var/cache/apk/* /var/cache/distfiles/*


