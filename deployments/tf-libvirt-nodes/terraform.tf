#
# Author(s): Alvaro Saurin <alvaro.saurin@suse.com>
#
# Copyright (c) 2017 SUSE LINUX GmbH, Nuernberg, Germany.
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative.
#

#####################
# Cluster variables #
#####################

variable "libvirt_uri" {
  default     = "qemu:///system"
  description = "libvirt connection url - default to localhost"
}

variable "img_pool" {
  default     = "default"
  description = "pool to be used to store all the volumes"
}

variable "img_url_base" {
  type        = "string"
  default     = "https://download.opensuse.org/repositories/devel:/kubic:/images:/experimental/images/"
  description = "base URL to the KVM image used (ie, we will download '<img_url_base>/<img_regex>.qcow2')"
}

variable "img_src_filename" {
  type        = "string"
  default     = ""
  description = "Force a specific filename"
}

variable "img" {
  type        = "string"
  default     = "../images/kubic.qcow2"
  description = "remote URL or local copy (can be used in conjuction with img_url_base) of the image to use."
}

variable "img_refresh" {
  default     = "true"
  description = "Try to get the latest image (true/false)"
}

variable "img_down_extra_args" {
  default     = ""
  description = "Extra arguments for the images downloader"
}

variable "img_regex" {
  type        = "string"
  default     = "kubeadm-cri-o-kvm-and-xen"
  description = "regex for selecting the image filename (ie, we will download '<img_url_base>/<img_regex>.qcow2')"
}

variable "nodes_count" {
  default     = 1
  description = "Number of non-seed nodes to be created"
}

variable "prefix" {
  type        = "string"
  default     = "kubic"
  description = "a prefix for resources"
}

variable "network" {
  type        = "string"
  default     = "default"
  description = "an existing network to use for the VMs"
}

variable "cni_driver" {
  type        = "string"
  default     = "flannel"
  description = "the CNI driver to use: flannel, cilium..."
}

variable "password" {
  type        = "string"
  default     = "linux"
  description = "password for sshing to the VMs"
}

variable "devel" {
  type        = "string"
  default     = "1"
  description = "enable some steps for development environments (non-empty=true)"
}

variable "kubic_init_image_name" {
  type        = "string"
  default     = "kubic-project/kubic-init:latest"
  description = "the default kubic init image name"
}

variable "kubic_init_image_tgz" {
  type        = "string"
  default     = "kubic-init-latest.tar.gz"
  description = "a kubic-init container image"
}

variable "kubic_init_runner" {
  type        = "string"
  default     = "podman"
  description = "the kubic-init runner: docker or podman"
}

variable "kubic_init_extra_args" {
  type        = "string"
  default     = ""
  description = "extra args for the kubic-init bootstrap"
}

variable "default_node_memory" {
  default     = 2048
  description = "default amount of RAM of the Nodes (in bytes)"
}

variable "nodes_memory" {
  default = {
    "3" = "1024"
    "4" = "1024"
    "5" = "1024"
  }

  description = "amount of RAM for some specific nodes"
}

data "template_file" "init_script" {
  template = "${file("../support/tf/init.sh.tpl")}"

  vars {
    kubic_init_image_name = "${var.kubic_init_image_name}"
    kubic_init_image_tgz  = "${var.kubic_init_image_tgz}"
    kubic_init_runner     = "${var.kubic_init_runner}"
    kubic_init_extra_args = "${var.kubic_init_extra_args}"
  }
}

#######################
# Cluster declaration #
#######################

provider "libvirt" {
  uri = "${var.libvirt_uri}"
}

#######################
# Base image          #
#######################

resource "null_resource" "download_kubic_image" {
  count = "${length(var.img_url_base) == 0 ? 0 : 1}"

  provisioner "local-exec" {
    command = "../support/tf/download-image.sh --img-regex '${var.img_regex}' --libvirt-uri '${var.libvirt_uri}' --src-base '${var.img_url_base}' --refresh '${var.img_refresh}' --local '${var.img}' --upload-to-img '${var.prefix}_base_${basename(var.img)}' --upload-to-pool '${var.img_pool}' --src-filename '${var.img_src_filename}' ${var.img_down_extra_args}"
  }
}

###########################
# Local IP (for seeding)  #
###########################

data "external" "seeder" {
  program = [
    "python",
    "../support/tf/get-seeder.py",
  ]
}

output "seeder_ip" {
  value = "${data.external.seeder.result.ip}"
}

###########################
# Token                   #
###########################

data "external" "token_get" {
  program = [
    "python",
    "../support/tf/get-token.py",
  ]
}

output "token" {
  value = "${data.external.token_get.result.token}"
}

###########################
# Cluster non-seed nodes #
###########################

resource "libvirt_volume" "node" {
  count            = "${var.nodes_count}"
  name             = "${var.prefix}_node_${count.index}.qcow2"
  pool             = "${var.img_pool}"
  base_volume_name = "${var.prefix}_base_${basename(var.img)}"

  depends_on = [
    "null_resource.download_kubic_image",
  ]
}

data "template_file" "node_cloud_init_user_data" {
  count    = "${var.nodes_count}"
  template = "${file("../cloud-init/node.cfg.tpl")}"

  vars {
    seeder     = "${data.external.seeder.result.ip}"
    token      = "${data.external.token_get.result.token}"
    password   = "${var.password}"
    hostname   = "${var.prefix}-node-${count.index}"
    cni_driver = "${var.cni_driver}"
  }
}

resource "libvirt_cloudinit_disk" "node" {
  count     = "${var.nodes_count}"
  name      = "${var.prefix}_node_cloud_init_${count.index}.iso"
  pool      = "${var.img_pool}"
  user_data = "${element(data.template_file.node_cloud_init_user_data.*.rendered, count.index)}"
}

resource "libvirt_domain" "node" {
  count     = "${var.nodes_count}"
  name      = "${var.prefix}-node-${count.index}"
  memory    = "${lookup(var.nodes_memory, count.index, var.default_node_memory)}"
  cloudinit = "${element(libvirt_cloudinit_disk.node.*.id, count.index)}"

  disk {
    volume_id = "${element(libvirt_volume.node.*.id, count.index)}"
  }

  network_interface {
    network_name   = "${var.network}"
    wait_for_lease = 1
  }

  graphics {
    type        = "vnc"
    listen_type = "address"
  }
}

resource "null_resource" "upload_config_nodes" {
  count = "${length(var.devel) == 0 ? 0 : var.nodes_count}"

  connection {
    host     = "${element(libvirt_domain.node.*.network_interface.0.addresses.0, count.index)}"
    password = "${var.password}"
  }

  provisioner "remote-exec" {
    inline = [
      "mkdir -p /etc/systemd/system/kubelet.service.d",
    ]
  }

  provisioner "file" {
    source      = "../../init/kubelet.drop-in.conf"
    destination = "/etc/systemd/system/kubelet.service.d/kubelet.conf"
  }

  provisioner "file" {
    source      = "../../init/kubic-init.systemd.conf"
    destination = "/etc/systemd/system/kubic-init.service"
  }

  provisioner "file" {
    source      = "../../init/kubic-init.sysconfig"
    destination = "/etc/sysconfig/kubic-init"
  }

  provisioner "file" {
    source      = "../../init/kubelet-sysctl.conf"
    destination = "/etc/sysctl.d/99-kubernetes-cri.conf"
  }

  provisioner "file" {
    source      = "../../${var.kubic_init_image_tgz}"
    destination = "/tmp/${var.kubic_init_image_tgz}"
  }

  # TODO: this is only for development
  provisioner "remote-exec" {
    inline = "${data.template_file.init_script.rendered}"
  }
}

output "nodes" {
  value = "${libvirt_domain.node.*.network_interface.0.addresses}"
}
