provider "kubernetes" {
    config_path = "~/.kube/config"
    config_context = "microk8s-cluster"
}

resource "kubernetes_namespace" "swiss-knife" {
    metadata {
        name = "swiss-knife"
    }
}

resource "kubernetes_service" "swiss-knife" {
    metadata {
        name = "swiss-knife"
        namespace = kubernetes_namespace.swiss-knife.metadata[0].name
    }

    spec {
        selector = {
            app = "swiss-knife"
        }

        port {
            port = 8080
            target_port = 8080
        }
    }
}


resource "kubernetes_persistent_volume_claim_v1" "swiss-knife" {
    metadata {
        name = "swiss-knife"
        namespace = kubernetes_namespace.swiss-knife.metadata[0].name
    }

    spec {
        access_modes = ["ReadWriteOnce"]
        resources {
            requests = {
                storage = "1Gi"
            }
        }
    }
}

resource "kubernetes_ingress_v1" "swiss-knife" {
    metadata {
        name = "swiss-knife"
        namespace = kubernetes_namespace.swiss-knife.metadata[0].name
        annotations = {
            "cert-manager.io/cluster-issuer" = "letsencrypt-prod"
            "acme.cert-manager.io/http01-edit-in-place"    = "false"
        }
    }

    spec {
        ingress_class_name = "nginx-private"
        tls {
            hosts       = ["swiss-knife.i.siriusfrk.ru"]
            secret_name = "swiss-knife-i-siriusfrk-ru-tls"
        }

        rule {

            host = "swiss-knife.i.siriusfrk.ru"
            http {
                path {
                    path = "/"
                    path_type = "Prefix"
                    backend {
                        service {
                            name = kubernetes_service.swiss-knife.metadata[0].name
                            port {
                                number = kubernetes_service.swiss-knife.spec[0].port[0].port
                            }
                        }
                    }
                }
            }
        }
    }
}


resource "kubernetes_deployment_v1" "swiss-knife" {
    metadata {
        name = "swiss-knife"
        namespace = kubernetes_namespace.swiss-knife.metadata[0].name
    }

    spec {
        replicas = 1

        selector {
            match_labels = {
                app = "swiss-knife"
            }
        }

        template {
            metadata {
                labels = {
                    app = "swiss-knife"
                }
            }

            spec {
                node_selector = {
                    "node.kubernetes.io/microk8s-worker" = "microk8s-worker"
                }

                container {
                    image = "registry.i.siriusfrk.ru/swiss-knife:latest"
                    name = "swiss-knife"
                    image_pull_policy = "Always"
                    port {
                        container_port = 8080
                    }
                    volume_mount {
                        name = "swiss-knife"
                        mount_path = "/data"
                    }
                }

                volume {
                    name = "swiss-knife"
                    persistent_volume_claim {
                        claim_name = kubernetes_persistent_volume_claim_v1.swiss-knife.metadata[0].name
                    }
                }
            }
        }
    }
}
