FROM mcr.microsoft.com/vscode/devcontainers/go
RUN go install go.uber.org/mock/mockgen@latest
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN go install github.com/dkorunic/betteralign/cmd/betteralign@latest
RUN sudo chown -R vscode:golang /go