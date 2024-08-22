FROM golang:1.19.8 AS build

COPY src /src
WORKDIR /src/mkgchat
RUN go build -o /src/bin/mkgchat -tags netgo -ldflags '-w -extldflags "-static"' main.go
WORKDIR /src/togchat
RUN go build -o /src/bin/togchat -tags netgo -ldflags '-w -extldflags "-static"' main.go
WORKDIR /src/eventdiff
RUN go build -o /src/bin/eventdiff -tags netgo -ldflags '-w -extldflags "-static"' main.go

FROM docker.io/dim9/static:combi AS final
WORKDIR /

COPY --from=build /src/bin/mkgchat /bin/mkgchat
COPY --from=build /src/bin/togchat /bin/togchat
COPY --from=build /src/bin/eventdiff /bin/eventdiff


CMD [ "/bin/bash" ]
