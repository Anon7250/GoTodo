FROM golang
COPY . /work
WORKDIR /work
RUN go build -o todo . && mv todo /todo && go clean -cache -testcache -modcache . && rm -rf --one-file-system /work
CMD /todo
