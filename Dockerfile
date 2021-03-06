# This is a multi-stage Dockerfile and requires >= Docker 17.05
# https://docs.docker.com/engine/userguide/eng-image/multistage-build/
FROM quay.io/slukasik/buffalo-builder as builder

RUN apt-get update && apt-get install -y libxml2-dev zlib1g-dev liblzma-dev libicu-dev

RUN GO111MODULE=off go get -u -v github.com/gocomply/fedramp/cli/gocomply_fedramp

RUN mkdir -p $GOPATH/src/github.com/RedHatGov/ocdb
WORKDIR $GOPATH/src/github.com/RedHatGov/ocdb
ENV GO111MODULE on
ENV GOPROXY http://proxy.golang.org

# this will cache the npm install step, unless package.json changes
ADD package.json .
ADD yarn.lock .
RUN yarn install --no-progress
ADD . .
RUN go get ./...
RUN buffalo build --ldflags '-linkmode external -extldflags "-static -lz -llzma -licuuc -licudata -ldl -lstdc++ -lm"' -o /bin/app

RUN mkdir -p /var/tmp/ocdb
WORKDIR /var/tmp/ocdb
RUN git clone --depth 1 https://github.com/ComplianceAsCode/oscal ComplianceAsCode.oscal
RUN cd ComplianceAsCode.oscal && PATH=$GOPATH:$PATH make docx

FROM registry.centos.org/centos:8

WORKDIR /bin/
COPY --from=builder /bin/app /go/bin/gocomply_fedramp ./

WORKDIR /var/tmp/
COPY --from=builder /var/tmp/ocdb ocdb

RUN \
	dnf install -y 'dnf-command(copr)' && \
	dnf copr enable -y openscapmaint/openscap-latest && \
	dnf install --setopt=tsflags=nodocs -y \
		bash git ca-certificates cmake make openscap-scanner python3-pyyaml python3-jinja2 python3 && \
	dnf clean all && \
  chmod --recursive og+w /var/tmp/ocdb

# Uncomment to run the binary in "production" mode:
# ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
ENV ADDR=0.0.0.0

EXPOSE 3000

# Uncomment to run the migrations before running the binary:
# CMD /bin/app migrate; /bin/app
CMD exec /bin/app
