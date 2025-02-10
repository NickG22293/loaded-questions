FROM mcr.microsoft.com/devcontainers/base:ubuntu

# Install Go
RUN curl -fsSL https://golang.org/dl/go1.21.1.linux-amd64.tar.gz | tar -xz -C /usr/local && \
    ln -s /usr/local/go/bin/go /usr/local/bin/go && \
    ln -s /usr/local/go/bin/gofmt /usr/local/bin/gofmt

# Install Node.js
RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && \
    apt-get install -y nodejs

# Install Terraform
RUN sudo apt-get update && sudo apt-get install -y gnupg software-properties-common && \
    wget -O- https://apt.releases.hashicorp.com/gpg | \ 
    gpg --dearmor | \ 
    sudo tee /usr/share/keyrings/hashicorp-archive-keyring.gpg > /dev/null && \
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] \
    https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
    sudo tee /etc/apt/sources.list.d/hashicorp.list && \
    sudo apt update && \
    sudo apt-get install terraform


# Install kubectl
RUN curl -fsSL -o /usr/local/bin/kubectl https://dl.k8s.io/release/v1.27.0/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl

# Install Docker CLI
RUN apt-get update && apt-get install -y docker.io

# Set default user (root for docker sock permissions)
USER root
