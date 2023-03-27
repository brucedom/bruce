variable "region" {
  type    = string
  default = "us-east-1"
}

variable "vpc_cidr" {
  type = string
}

variable "name" {
  type = string
  default = "bruce-test"
}

variable "environment" {
  type = string
}

variable "ec2-ami"{
  type = string
}

variable "ec2-type"{
  type = string
  default = "t2.micro"
}

variable "key-name"{
  type = string
  default = "Nitecon"
}
variable "createPublicIP"{
  type = bool
  default = true
}
variable "max-subnets"{
  type = number
  default = 1
}