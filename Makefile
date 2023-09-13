# Makefile для сборки и установки агента и сервера

# Компилятор Go
GO := go

# Параметры сборки агента
AGENT_SRC := cmd/agent/main.go
AGENT_BINARY := cmd/agent/agent
AGENT_OUTPUT := cmd/agent/agent

# Параметры сборки сервера
SERVER_SRC := cmd/server/main.go
SERVER_BINARY := cmd/server/server
SERVER_OUTPUT := cmd/server/server

# Цель по умолчанию: собрать агент и сервер
all: agent server

# Цель для сборки агента
agent:
	$(GO) build -o $(AGENT_OUTPUT) $(AGENT_SRC)

# Цель для сборки сервера
server:
	$(GO) build -o $(SERVER_OUTPUT) $(SERVER_SRC)

# Цель для очистки скомпилированных файлов
clean:
	rm -f $(AGENT_OUTPUT) $(SERVER_OUTPUT)

.PHONY: all agent server clean"