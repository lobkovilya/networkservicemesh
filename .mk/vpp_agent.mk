ARCH ?= $(shell uname -m)
ifeq (${ARCH}, x86_64)
  export VPP_AGENT=ligato/vpp-agent:dev
  export VPP_AGENT_DEV=ligato/dev-vpp-agent:dev
endif
ifeq (${ARCH}, aarch64)
  export VPP_AGENT=ligato/vpp-agent-arm64:v2.0.2
  export VPP_AGENT_DEV=ligato/dev-vpp-agent-arm64:v2.0.2
endif
