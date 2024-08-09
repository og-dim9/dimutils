FROM docker.io/dim9/static:bash AS bash
FROM docker.io/dim9/static:jq AS jq
FROM docker.io/dim9/static:kcat AS kcat
FROM docker.io/dim9/static:make AS make
FROM docker.io/dim9/static:node AS node
FROM docker.io/dim9/static:opentofu AS opentofu
FROM docker.io/dim9/static:quickjs AS quickjs
FROM docker.io/dim9/static:task AS task
FROM docker.io/dim9/static:terraform AS terraform
FROM docker.io/dim9/static:yq AS yq
FROM docker.io/dim9/static:busybox AS final
COPY --from=bash /bin/bash /bin/
COPY --from=jq /bin/jq /bin/
COPY --from=kcat /bin/kcat /bin/
COPY --from=make /bin/make /bin/
COPY --from=node /bin/node /bin/
COPY --from=opentofu /bin/opentofu /bin/
COPY --from=quickjs /bin/quickjs /bin/
COPY --from=task /bin/task /bin/
COPY --from=terraform /bin/terraform /bin/
COPY --from=yq /bin/yq /bin/

USER dimutls

CMD [ "/bin/bash" ]
