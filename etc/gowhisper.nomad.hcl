# Gowhisper Job template
# Variables:
#
#   dc                 - list(string): Data centers for job
#   namespace          - string: Namespace of the job (default "default")
#   hosts              - list(string): Hostname constraint to run on (optional)
#   port               - number: External port (default 8081)
#   memory             - number: Memory allocation in MB (default 4096)
#   devices            - list(string): Devices to expose (optional)
#   data               - string: Host path for media storage (optional)
#   service_name       - string: Service name for Nomad and OTEL (default "gowhisper")
#   service_dns        - list(string): DNS servers (optional)
#   docker_image       - string: Docker image (default "ghcr.io/mutablelogic/go-whisper")
#   docker_runtime     - string: Docker runtime (optional)
#   docker_user        - string: Docker registry username (optional)
#   docker_token       - string: Docker registry access token (optional)
#   openai_api_key     - string: OpenAI API Key (optional)
#   elevenlabs_api_key - string: ElevenLabs API Key (optional)
#   whisper            - object: Whisper configuration (max_contexts, gpu)
#   segmenter          - object: Segmenter configuration (min_silence_size, max_segment_size)
#   otel_endpoint      - string: OTEL exporter endpoint (optional)
#   debug              - bool: Enable debug output (default false)

##########################################################################
# VARIABLES

variable "dc" {
  description = "data centers that the job runs in"
  type        = list(string)
}

variable "namespace" {
  description = "namespace that the job runs in"
  type        = string
  default     = "default"
}

variable "service_dns" {
  description = "dns servers"
  type        = list(string)
  default     = []
}

variable "hosts" {
  description = "hosts where job needs to be run"
  type        = list(string)
  default     = []
}

variable "docker_image" {
  description = "Docker image for gowhisper"
  type        = string
  default     = "ghcr.io/mutablelogic/go-whisper"
}

variable "docker_runtime" {
  description = "Docker runtime"
  type        = string
  default     = ""
}

variable "service_name" {
  description = "Service name for Nomad and OTEL"
  type        = string
  default     = "gowhisper"
}

variable "docker_user" {
  description = "Docker registry username"
  type        = string
  default     = ""
}

variable "docker_token" {
  description = "Docker registry access token"
  type        = string
  default     = ""
}

variable "port" {
  description = "External port"
  type        = number
  default     = 8081
}

variable "memory" {
  description = "Memory allocation in MB"
  type        = number
  default     = 4096
}

variable "devices" {
  description = "Devices to expose"
  type        = list(string)
  default     = []
}

variable "debug" {
  description = "Enable debug output"
  type        = bool
  default     = false
}

variable "otel_endpoint" {
  description = "OTEL exporter endpoint"
  type        = string
  default     = ""
}

variable "data" {
  description = "Host path for media storage (optional)"
  type        = string
  default     = ""
}


variable "openai_api_key" {
  description = "OpenAI API Key"
  type        = string
  default     = ""
}

variable "elevenlabs_api_key" {
  description = "ElevenLabs API Key"
  type        = string
  default     = ""
}

variable "whisper" {
  description = "Whisper configuration"
  type = object({
    max_contexts = number
    gpu          = bool
  })
  default = {
    max_contexts = 0
    gpu          = true
  }
}

variable "segmenter" {
  description = "Segmenter configuration"
  type        = object({
    min_silence_size = string
    max_segment_size = string
  })
  default = {
    min_silence_size = "500ms"
    max_segment_size = "15m"
  }
}

##########################################################################
# LOCALS

locals {
  volumes = compact([
    var.data != "" ? format("%s:/data", var.data) : null,
  ])
  devices = [
    for device in var.devices : {
      host_path      = device
      container_path = device
    }
  ]
}

##########################################################################
# JOB

job "gowhisper" {
  type        = "service"
  namespace   = var.namespace
  datacenters = var.dc

  dynamic "constraint" {
    for_each = var.hosts
    content {
      attribute = "${node.unique.name}"
      value     = constraint.value
    }
  }

  update {
    min_healthy_time  = "10s"
    healthy_deadline  = "10m"
    progress_deadline = "20m"
    health_check      = "task_states"
  }

  group "gowhisper" {
    count = 1

    network {
      port "http" {
        static = var.port
        to     = 8081
      }
    }

    service {
      name     = var.service_name
      port     = "http"
      tags     = ["http", var.service_name]
      provider = "nomad"
    }

    task "server" {
      driver = "docker"

      resources {
        memory = var.memory
      }

      env {
        GOWHISPER_ADDR              = "0.0.0.0:8081"
        OTEL_EXPORTER_OTLP_ENDPOINT = var.otel_endpoint
        OTEL_SERVICE_NAME           = var.service_name
        OPENAI_API_KEY              = var.openai_api_key
        ELEVENLABS_API_KEY          = var.elevenlabs_api_key
      }

      config {
        runtime     = var.docker_runtime
        image       = var.docker_image
        force_pull  = true
        dns_servers = var.service_dns
        ports       = ["http"]
        volumes     = local.volumes
        devices     = local.devices

        auth {
          username = var.docker_user
          password = var.docker_token
        }

        args = concat(
          ["run"],
          var.debug ? ["--debug"] : [],
          var.data != "" ? ["--models", "/data"] : [],
          var.whisper.max_contexts != 0 ? ["--whisper.max-contexts", format("%d", var.whisper.max_contexts)] : [],
          var.whisper.gpu ? ["--whisper.gpu=true"] : ["--whisper.gpu=false"],
          var.segmenter.min_silence_size != "" ? ["--segmenter.min-silence-size", var.segmenter.min_silence_size] : [],
          var.segmenter.max_segment_size != "" ? ["--segmenter.max-segment-size", var.segmenter.max_segment_size] : [],
        )
      }
    }
  }
}
