FROM golang:1.14

ENV GO111MODULE=on
ENV GOFLAGS=-mod=readonly
ENV APP_USER app
ENV APP_HOME /app

ARG GID
ARG UID

RUN groupadd --gid $GID $APP_USER && useradd --uid $UID --gid $GID -m $APP_USER
RUN mkdir -p $APP_HOME && chown -R $APP_USER:$APP_USER $APP_HOME

USER $APP_USER
WORKDIR $APP_HOME
COPY . .
RUN go build $GOFLAGS
EXPOSE 8000
CMD ["./tsatsubii-auth-service"]
