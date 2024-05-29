FROM busybox:musl as builder

ARG TASK_URL=https://github.com/go-task/task/releases/download
ARG TASK_VERSION=v3.37.2
ARG TASK_ARCH=linux_amd64

WORKDIR /install/task
RUN wget --no-check-certificate \
    "${TASK_URL}/${TASK_VERSION}/task_${TASK_ARCH}.tar.gz" && \
    tar -xzf task_${TASK_ARCH}.tar.gz  task -C /bin && \
    rm -rfv task_${TASK_ARCH}.tar.gz

# FROM debian:stable-slim as bash
# RUN apt update && apt intall -y bash-static
# RUN which bash-static

FROM scratch as final
COPY --from=builder /bin/busybox /bin/busybox
COPY --from=builder /bin/task /bin/task
# COPY --from=bash /usr/bin/bash-static /bin/bash
RUN ["/bin/busybox", "--install", "/bin"]
#COPY --from=builder /app/bin/task /usr/local/bin/task
