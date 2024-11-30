# Pinned major verison of 1 and will update to latest 1.X.X
ARG base_image=liveoak/ci-pipeline/alpine
ARG base_version=1
FROM registry.gitlab.com/${base_image}:${base_version}

# Install Project Dependencies
RUN install-python
RUN install-aws-cli
RUN install-aws-sso-util
RUN install-terraform-1.7
RUN install-go
RUN install-pre-commit
RUN install-zsh

RUN apk add neofetch

# ZSH Config
ENV SHELL=/bin/zsh
COPY zshrc /root/.zshrc
